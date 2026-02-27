DROP INDEX IF EXISTS idx_audit_logs_timestamp_desc;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash_unique;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS timestamp;

ALTER TABLE refresh_tokens
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS replaced_by_token_id,
    DROP COLUMN IF EXISTS issued_at;
