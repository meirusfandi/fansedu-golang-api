-- Review manual per jawaban (admin / trainer bidang terkait): komentar + skor override.

ALTER TABLE attempt_answers
  ADD COLUMN IF NOT EXISTS reviewer_comment TEXT,
  ADD COLUMN IF NOT EXISTS manual_score DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS reviewed_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_attempt_answers_reviewed_at ON attempt_answers (reviewed_at) WHERE reviewed_at IS NOT NULL;
