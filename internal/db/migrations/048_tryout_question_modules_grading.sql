-- Modul/topik + kunci jawaban untuk analisis per soal; is_correct di attempt_answers setelah submit.

ALTER TABLE questions
  ADD COLUMN IF NOT EXISTS module_id TEXT,
  ADD COLUMN IF NOT EXISTS module_title TEXT,
  ADD COLUMN IF NOT EXISTS bidang TEXT,
  ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS correct_option VARCHAR(50),
  ADD COLUMN IF NOT EXISTS correct_text TEXT;

ALTER TABLE attempt_answers
  ADD COLUMN IF NOT EXISTS is_correct BOOLEAN;

CREATE INDEX IF NOT EXISTS idx_attempt_answers_attempt_correct ON attempt_answers (attempt_id) WHERE is_correct IS NOT NULL;
