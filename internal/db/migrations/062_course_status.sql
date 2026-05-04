-- Add lifecycle status for admin-managed courses.
-- draft   : editable/incomplete, hidden from public course endpoints
-- publish : visible in catalog (ready to review)
-- active  : actively promoted/open for enrollment

ALTER TABLE courses
  ADD COLUMN IF NOT EXISTS status VARCHAR(20);

UPDATE courses
SET status = 'active'
WHERE status IS NULL OR BTRIM(status) = '';

ALTER TABLE courses
  ALTER COLUMN status SET DEFAULT 'draft';

ALTER TABLE courses
  ALTER COLUMN status SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_courses_status'
  ) THEN
    ALTER TABLE courses
      ADD CONSTRAINT chk_courses_status
      CHECK (status IN ('draft', 'publish', 'active'));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_courses_status ON courses (status);
