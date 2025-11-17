ALTER TABLE users
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
ADD COLUMN verified_at TIMESTAMP;

UPDATE users SET is_verified = TRUE WHERE is_verified = FALSE OR is_verified IS NULL;
UPDATE users SET verified_at = CURRENT_TIMESTAMP WHERE verified_at IS NULL AND is_verified = TRUE;

CREATE TABLE IF NOT EXISTS verification_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_verification_tokens_token ON verification_tokens(token);
