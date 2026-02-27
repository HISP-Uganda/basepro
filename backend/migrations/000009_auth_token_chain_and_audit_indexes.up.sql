ALTER TABLE refresh_tokens
    ADD COLUMN IF NOT EXISTS issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS replaced_by_token_id BIGINT REFERENCES refresh_tokens(id),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash_unique ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp_desc ON audit_logs(timestamp DESC);
