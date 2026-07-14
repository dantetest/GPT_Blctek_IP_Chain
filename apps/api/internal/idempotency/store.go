package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/id"
	"github.com/go-sql-driver/mysql"
)

type Record struct {
	ID             string
	RequestHash    string
	ResponseStatus *int
	ResponseBody   []byte
	ExpiresAt      time.Time
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Begin(ctx context.Context, scope, key, requestHash string, expiresAt time.Time) (Record, bool, error) {
	now := time.Now().UTC()
	if record, err := s.find(ctx, scope, key); err == nil {
		if record.ExpiresAt.After(now) {
			return record, false, nil
		}
		if _, err := s.db.ExecContext(ctx, `DELETE FROM idempotency_records WHERE id=?`, record.ID); err != nil {
			return Record{}, false, err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Record{}, false, err
	}

	recordID, err := id.New()
	if err != nil {
		return Record{}, false, err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO idempotency_records (
			id,scope,idempotency_key,request_hash,expires_at,created_at
		) VALUES (?,?,?,?,?,?)`,
		recordID,
		scope,
		key,
		requestHash,
		expiresAt,
		now,
	)
	if err == nil {
		return Record{ID: recordID, RequestHash: requestHash, ExpiresAt: expiresAt}, true, nil
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		record, findErr := s.find(ctx, scope, key)
		return record, false, findErr
	}
	return Record{}, false, fmt.Errorf("create idempotency record: %w", err)
}

func (s *Store) Complete(ctx context.Context, recordID string, status int, body []byte) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE idempotency_records
		SET response_status=?,response_body=?
		WHERE id=?`, status, body, recordID)
	return err
}

func (s *Store) Delete(ctx context.Context, recordID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM idempotency_records WHERE id=?`, recordID)
	return err
}

func (s *Store) find(ctx context.Context, scope, key string) (Record, error) {
	var record Record
	var responseStatus sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT id,request_hash,response_status,response_body,expires_at
		FROM idempotency_records WHERE scope=? AND idempotency_key=?`, scope, key).Scan(
		&record.ID,
		&record.RequestHash,
		&responseStatus,
		&record.ResponseBody,
		&record.ExpiresAt,
	)
	if responseStatus.Valid {
		status := int(responseStatus.Int64)
		record.ResponseStatus = &status
	}
	return record, err
}
