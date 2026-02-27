ALTER TABLE api_tokens
    ADD COLUMN IF NOT EXISTS prefix TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'api_tokens' AND column_name = 'user_id'
    ) THEN
        UPDATE api_tokens
        SET created_by_user_id = COALESCE(created_by_user_id, user_id)
        WHERE created_by_user_id IS NULL;
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS api_token_permissions (
    id BIGSERIAL PRIMARY KEY,
    api_token_id BIGINT NOT NULL REFERENCES api_tokens(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    module_scope TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_token_permissions_unique
    ON api_token_permissions (api_token_id, permission, module_scope);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_token_hash_unique ON api_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_api_tokens_revoked_at ON api_tokens(revoked_at);
CREATE INDEX IF NOT EXISTS idx_api_tokens_prefix ON api_tokens(prefix);
