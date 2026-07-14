package apikey

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type UserEntitlement struct {
	Status string
	Tier   string
	Role   string
}

func (r *Repository) UserEntitlement(ctx context.Context, userID string) (UserEntitlement, error) {
	var entitlement UserEntitlement
	err := r.db.QueryRowContext(ctx, `
		SELECT status,tier,global_role FROM users WHERE id=?`, userID).Scan(
		&entitlement.Status,
		&entitlement.Tier,
		&entitlement.Role,
	)
	return entitlement, err
}

func (r *Repository) CreateWithLimit(ctx context.Context, key Key, limit *int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=? FOR UPDATE`, key.UserID).Scan(&userID); err != nil {
		return err
	}
	if limit != nil {
		var count int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM api_keys WHERE user_id=? AND status='ACTIVE'`, key.UserID).Scan(&count); err != nil {
			return err
		}
		if count >= *limit {
			return ErrLimitReached
		}
	}

	scopes, err := json.Marshal(key.Scopes)
	if err != nil {
		return fmt.Errorf("marshal api key scopes: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO api_keys (
			id,user_id,name,key_prefix,secret_hash,scopes,status,monthly_quota,
			monthly_used,quota_month,expires_at,created_at
		) VALUES (?,?,?,?,?,?,'ACTIVE',?,0,?,?,?)`,
		key.ID,
		key.UserID,
		key.Name,
		key.KeyPrefix,
		key.SecretHash,
		scopes,
		key.MonthlyQuota,
		key.QuotaMonth,
		key.ExpiresAt,
		key.CreatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) List(ctx context.Context, userID string) ([]Key, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,user_id,name,key_prefix,secret_hash,scopes,status,monthly_quota,
			monthly_used,quota_month,expires_at,last_used_at,created_at,revoked_at
		FROM api_keys WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]Key, 0)
	for rows.Next() {
		key, err := scanKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r *Repository) Revoke(ctx context.Context, userID, keyID string, at time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE api_keys SET status='REVOKED',revoked_at=?
		WHERE id=? AND user_id=? AND status='ACTIVE'`, at, keyID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrKeyNotFound
	}
	return nil
}

func (r *Repository) AuthenticateAndConsume(ctx context.Context, raw string, now time.Time) (Principal, error) {
	prefix, err := Parse(raw)
	if err != nil {
		return Principal{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Principal{}, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		SELECT k.id,k.user_id,k.secret_hash,k.scopes,k.status,k.monthly_quota,k.monthly_used,
			k.quota_month,k.expires_at,u.status,u.tier,u.global_role
		FROM api_keys k
		JOIN users u ON u.id=k.user_id
		WHERE k.key_prefix=? FOR UPDATE`, prefix)
	var principal Principal
	var secretHash string
	var scopesJSON []byte
	var keyStatus, userStatus string
	var monthlyQuota *int64
	var monthlyUsed int64
	var quotaMonth time.Time
	err = row.Scan(
		&principal.KeyID,
		&principal.UserID,
		&secretHash,
		&scopesJSON,
		&keyStatus,
		&monthlyQuota,
		&monthlyUsed,
		&quotaMonth,
		&principal.ExpiresAt,
		&userStatus,
		&principal.Tier,
		&principal.Role,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Principal{}, ErrInvalidKey
	}
	if err != nil {
		return Principal{}, err
	}
	if keyStatus != StatusActive || subtle.ConstantTimeCompare([]byte(secretHash), []byte(Hash(raw))) != 1 {
		return Principal{}, ErrInvalidKey
	}
	if userStatus != "ACTIVE" {
		return Principal{}, ErrAccountNotActive
	}
	if principal.ExpiresAt != nil && !principal.ExpiresAt.After(now) {
		return Principal{}, ErrInvalidKey
	}
	if err := json.Unmarshal(scopesJSON, &principal.Scopes); err != nil {
		return Principal{}, fmt.Errorf("unmarshal api key scopes: %w", err)
	}

	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if quotaMonth.Before(currentMonth) {
		monthlyUsed = 0
		quotaMonth = currentMonth
	}
	if monthlyQuota != nil && monthlyUsed >= *monthlyQuota {
		return Principal{}, ErrQuotaExceeded
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE api_keys SET monthly_used=?,quota_month=?,last_used_at=? WHERE id=?`,
		monthlyUsed+1,
		quotaMonth,
		now,
		principal.KeyID,
	)
	if err != nil {
		return Principal{}, err
	}
	if err := tx.Commit(); err != nil {
		return Principal{}, err
	}
	return principal, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanKey(row rowScanner) (Key, error) {
	var key Key
	var scopesJSON []byte
	err := row.Scan(
		&key.ID,
		&key.UserID,
		&key.Name,
		&key.KeyPrefix,
		&key.SecretHash,
		&scopesJSON,
		&key.Status,
		&key.MonthlyQuota,
		&key.MonthlyUsed,
		&key.QuotaMonth,
		&key.ExpiresAt,
		&key.LastUsedAt,
		&key.CreatedAt,
		&key.RevokedAt,
	)
	if err != nil {
		return Key{}, err
	}
	if err := json.Unmarshal(scopesJSON, &key.Scopes); err != nil {
		return Key{}, err
	}
	return key, nil
}
