-- +goose Up
ALTER TABLE users
  ADD COLUMN global_role ENUM('USER','ADMIN') NOT NULL DEFAULT 'USER' AFTER tier,
  ADD COLUMN failed_login_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0 AFTER email_verified_at,
  ADD COLUMN locked_until DATETIME(6) NULL AFTER failed_login_attempts,
  ADD COLUMN last_login_at DATETIME(6) NULL AFTER locked_until,
  ADD KEY idx_users_status_lock (status,locked_until);

CREATE TABLE email_verification_tokens (
  id VARCHAR(26) PRIMARY KEY,
  user_id VARCHAR(26) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  used_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY uk_email_verification_token (token_hash),
  KEY idx_email_verification_user (user_id,created_at),
  KEY idx_email_verification_expiry (expires_at),
  CONSTRAINT fk_email_verification_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE auth_sessions
  ADD KEY idx_auth_sessions_expiry (expires_at),
  ADD KEY idx_auth_sessions_active_user (user_id,revoked_at,expires_at);

-- +goose Down
ALTER TABLE auth_sessions
  DROP KEY idx_auth_sessions_active_user,
  DROP KEY idx_auth_sessions_expiry;

DROP TABLE IF EXISTS email_verification_tokens;

ALTER TABLE users
  DROP KEY idx_users_status_lock,
  DROP COLUMN last_login_at,
  DROP COLUMN locked_until,
  DROP COLUMN failed_login_attempts,
  DROP COLUMN global_role;
