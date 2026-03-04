DROP INDEX IF EXISTS users_email_unique_idx;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS telegram_handle,
    DROP COLUMN IF EXISTS whatsapp_number,
    DROP COLUMN IF EXISTS phone_number,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name,
    DROP COLUMN IF EXISTS language,
    DROP COLUMN IF EXISTS email;
