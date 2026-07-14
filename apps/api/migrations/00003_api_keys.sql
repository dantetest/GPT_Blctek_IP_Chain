-- +goose Up
CREATE TABLE api_keys (
  id VARCHAR(26) PRIMARY KEY,
  user_id VARCHAR(26) NOT NULL,
  name VARCHAR(120) NOT NULL,
  key_prefix VARCHAR(12) NOT NULL,
  secret_hash CHAR(64) NOT NULL,
  scopes JSON NOT NULL,
  status ENUM('ACTIVE','REVOKED') NOT NULL DEFAULT 'ACTIVE',
  monthly_quota BIGINT UNSIGNED NULL,
  monthly_used BIGINT UNSIGNED NOT NULL DEFAULT 0,
  quota_month DATE NOT NULL,
  expires_at DATETIME(6) NULL,
  last_used_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  revoked_at DATETIME(6) NULL,
  UNIQUE KEY uk_api_keys_prefix (key_prefix),
  UNIQUE KEY uk_api_keys_secret_hash (secret_hash),
  KEY idx_api_keys_user_status (user_id,status,created_at),
  KEY idx_api_keys_expiry (status,expires_at),
  CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE idempotency_records
  MODIFY response_body JSON NULL,
  ADD KEY idx_idempotency_expiry (expires_at);

-- +goose Down
ALTER TABLE idempotency_records
  DROP KEY idx_idempotency_expiry;

DROP TABLE IF EXISTS api_keys;
