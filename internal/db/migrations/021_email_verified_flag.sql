-- Add boolean email_verified flag and backfill existing users

ALTER TABLE users
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Backfill: mark all existing users as verified (and set timestamp if missing)
UPDATE users
SET
  email_verified = TRUE,
  email_verified_at = COALESCE(email_verified_at, NOW())
WHERE email_verified = FALSE;

