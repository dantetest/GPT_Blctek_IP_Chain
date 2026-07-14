package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateUserParams struct {
	User               User
	VerificationID     string
	VerificationHash   string
	VerificationToken  string
	VerificationExpiry time.Time
	OutboxEventID      string
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams, autoVerify bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin register transaction: %w", err)
	}
	defer tx.Rollback()

	var verifiedAt any
	if params.User.EmailVerifiedAt != nil {
		verifiedAt = *params.User.EmailVerifiedAt
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (
			id,email,password_hash,display_name,status,tier,global_role,email_verified_at,created_at,updated_at
		) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		params.User.ID,
		params.User.Email,
		params.User.PasswordHash,
		params.User.DisplayName,
		params.User.Status,
		params.User.Tier,
		params.User.GlobalRole,
		verifiedAt,
		params.User.CreatedAt,
		params.User.CreatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrEmailExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	if !autoVerify {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO email_verification_tokens (id,user_id,token_hash,expires_at,created_at)
			VALUES (?,?,?,?,?)`,
			params.VerificationID,
			params.User.ID,
			params.VerificationHash,
			params.VerificationExpiry,
			params.User.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert email verification token: %w", err)
		}

		payload, err := json.Marshal(map[string]any{
			"user_id": params.User.ID,
			"email":   params.User.Email,
			"token":   params.VerificationToken,
		})
		if err != nil {
			return fmt.Errorf("marshal verification event: %w", err)
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO outbox_events (
				id,aggregate_type,aggregate_id,event_type,payload,status,available_at,created_at
			) VALUES (?,?,?,?,?,'PENDING',?,?)`,
			params.OutboxEventID,
			"USER",
			params.User.ID,
			"USER_EMAIL_VERIFICATION_REQUESTED",
			payload,
			params.User.CreatedAt,
			params.User.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert verification outbox event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit register transaction: %w", err)
	}
	return nil
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx, userSelect+" WHERE email = ?", email))
}

func (r *Repository) FindUserByID(ctx context.Context, userID string) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx, userSelect+" WHERE id = ?", userID))
}

func (r *Repository) UnlockUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET status='ACTIVE', failed_login_attempts=0, locked_until=NULL
		WHERE id=? AND status='LOCKED'`, userID)
	return err
}

func (r *Repository) RecordLoginFailure(ctx context.Context, user User, maxAttempts int, lockUntil time.Time) error {
	attempts := user.FailedLoginAttempts + 1
	if attempts >= maxAttempts {
		_, err := r.db.ExecContext(ctx, `
			UPDATE users
			SET status='LOCKED', failed_login_attempts=?, locked_until=?
			WHERE id=?`, attempts, lockUntil, user.ID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE users SET failed_login_attempts=? WHERE id=?`, attempts, user.ID)
	return err
}

func (r *Repository) RecordLoginSuccess(ctx context.Context, userID string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET failed_login_attempts=0, locked_until=NULL, last_login_at=?,
			status=CASE WHEN status='LOCKED' THEN 'ACTIVE' ELSE status END
		WHERE id=?`, at, userID)
	return err
}

func (r *Repository) CreateSession(ctx context.Context, session Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (
			id,user_id,refresh_token_hash,user_agent,ip_address,expires_at,created_at
		) VALUES (?,?,?,?,?,?,?)`,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		time.Now().UTC(),
	)
	return err
}

func (r *Repository) FindSessionWithUser(ctx context.Context, refreshHash string) (Session, User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			s.id,s.user_id,s.refresh_token_hash,s.user_agent,s.ip_address,s.expires_at,s.revoked_at,
			u.id,u.email,u.password_hash,u.display_name,u.status,u.tier,u.global_role,
			u.email_verified_at,u.failed_login_attempts,u.locked_until,u.last_login_at,u.created_at
		FROM auth_sessions s
		JOIN users u ON u.id=s.user_id
		WHERE s.refresh_token_hash=?`, refreshHash)
	var session Session
	var user User
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.RevokedAt,
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Status,
		&user.Tier,
		&user.GlobalRole,
		&user.EmailVerifiedAt,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.LastLoginAt,
		&user.CreatedAt,
	)
	return session, user, err
}

func (r *Repository) RotateSession(ctx context.Context, oldSessionID string, session Session) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE auth_sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL`, time.Now().UTC(), oldSessionID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return ErrInvalidToken
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO auth_sessions (
			id,user_id,refresh_token_hash,user_agent,ip_address,expires_at,created_at
		) VALUES (?,?,?,?,?,?,?)`,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) RevokeSessionByRefreshHash(ctx context.Context, refreshHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE auth_sessions SET revoked_at=COALESCE(revoked_at,?) WHERE refresh_token_hash=?`,
		time.Now().UTC(),
		refreshHash,
	)
	return err
}

func (r *Repository) VerifyEmail(ctx context.Context, tokenHash string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var tokenID, userID string
	var expiresAt time.Time
	var usedAt *time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT id,user_id,expires_at,used_at
		FROM email_verification_tokens WHERE token_hash=? FOR UPDATE`, tokenHash).Scan(
		&tokenID,
		&userID,
		&expiresAt,
		&usedAt,
	)
	if errors.Is(err, sql.ErrNoRows) || usedAt != nil {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	if !expiresAt.After(now) {
		return ErrExpiredToken
	}
	if _, err := tx.ExecContext(ctx, `UPDATE email_verification_tokens SET used_at=? WHERE id=?`, now, tokenID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET status='ACTIVE',email_verified_at=? WHERE id=? AND status='PENDING_EMAIL'`, now, userID); err != nil {
		return err
	}
	return tx.Commit()
}

const userSelect = `
	SELECT id,email,password_hash,display_name,status,tier,global_role,email_verified_at,
		failed_login_attempts,locked_until,last_login_at,created_at
	FROM users`

func scanUser(row *sql.Row) (User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Status,
		&user.Tier,
		&user.GlobalRole,
		&user.EmailVerifiedAt,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.LastLoginAt,
		&user.CreatedAt,
	)
	return user, err
}
