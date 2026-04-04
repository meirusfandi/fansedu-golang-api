-- Format kelas admin: track meetings vs tryout, pertemuan 1–8 (PDF, PR, live), pre-test → tryout_session.

ALTER TABLE courses
  ADD COLUMN IF NOT EXISTS track_type VARCHAR(20) NOT NULL DEFAULT 'meetings'
  CHECK (track_type IN ('meetings', 'tryout'));

COMMENT ON COLUMN courses.track_type IS 'meetings = modul pertemuan (PDF/PR/live); tryout = fokus latihan dari course_tryouts';

CREATE TABLE IF NOT EXISTS course_meetings (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id       UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  meeting_number  INT NOT NULL CHECK (meeting_number >= 1 AND meeting_number <= 8),
  title           VARCHAR(500) NOT NULL DEFAULT '',
  detail_text     TEXT,
  pdf_url         TEXT,
  pr_title        VARCHAR(500),
  pr_description  TEXT,
  live_class_url  TEXT,
  sort_order      INT NOT NULL DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (course_id, meeting_number)
);

CREATE INDEX IF NOT EXISTS idx_course_meetings_course ON course_meetings (course_id);

-- Satu pre-test (tryout) opsional per kelas — soal latihan sebelum/di awal materi.
CREATE TABLE IF NOT EXISTS course_pretests (
  course_id          UUID PRIMARY KEY REFERENCES courses (id) ON DELETE CASCADE,
  tryout_session_id  UUID NOT NULL REFERENCES tryout_sessions (id) ON DELETE CASCADE
);

ALTER TABLE learning_lessons
  ADD COLUMN IF NOT EXISTS tryout_session_id UUID REFERENCES tryout_sessions (id) ON DELETE SET NULL;

COMMENT ON COLUMN learning_lessons.tryout_session_id IS 'Untuk lesson tipe quiz = link ke tryout (pre-test / latihan)';
