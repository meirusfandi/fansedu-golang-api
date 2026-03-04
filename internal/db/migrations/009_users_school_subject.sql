-- Add school_id and subject_id to users (optional: user's school and subject/bidang)

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS school_id UUID REFERENCES schools (id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_users_school ON users (school_id);
CREATE INDEX IF NOT EXISTS idx_users_subject ON users (subject_id);
