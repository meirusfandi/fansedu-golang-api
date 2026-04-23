-- Add explicit tryout subject and school level metadata for admin filtering.
ALTER TABLE tryout_sessions
  ADD COLUMN IF NOT EXISTS subject VARCHAR(100),
  ADD COLUMN IF NOT EXISTS school_level VARCHAR(20);

CREATE INDEX IF NOT EXISTS idx_tryout_sessions_subject_text ON tryout_sessions (subject);
CREATE INDEX IF NOT EXISTS idx_tryout_sessions_school_level ON tryout_sessions (school_level);
