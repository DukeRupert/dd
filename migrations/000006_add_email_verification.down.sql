-- Drop email verifications table and index
DROP TABLE IF EXISTS email_verifications;

-- Remove verified_at from users
ALTER TABLE users DROP COLUMN verified_at;