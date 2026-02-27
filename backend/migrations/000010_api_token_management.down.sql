DROP INDEX IF EXISTS idx_api_tokens_prefix;
DROP INDEX IF EXISTS idx_api_tokens_revoked_at;
DROP INDEX IF EXISTS idx_api_tokens_token_hash_unique;
DROP INDEX IF EXISTS idx_api_token_permissions_unique;

DROP TABLE IF EXISTS api_token_permissions;

ALTER TABLE api_tokens
    DROP COLUMN IF EXISTS last_used_at,
    DROP COLUMN IF EXISTS revoked_at,
    DROP COLUMN IF EXISTS created_by_user_id,
    DROP COLUMN IF EXISTS prefix;
