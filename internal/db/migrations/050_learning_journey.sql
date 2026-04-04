-- Learning journey: sections → lessons → per-user progress (live class + PDF materi, bukan video).

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'learning_lesson_type') THEN
    CREATE TYPE learning_lesson_type AS ENUM ('video', 'text', 'quiz', 'assignment');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS course_sections (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id   UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  title       VARCHAR(500) NOT NULL,
  sort_order  INT NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_course_sections_course ON course_sections (course_id);

CREATE TABLE IF NOT EXISTS learning_lessons (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  section_id       UUID NOT NULL REFERENCES course_sections (id) ON DELETE CASCADE,
  type             learning_lesson_type NOT NULL DEFAULT 'text',
  title            VARCHAR(500) NOT NULL,
  sort_order       INT NOT NULL DEFAULT 0,
  content          TEXT,
  pdf_url          TEXT,
  live_class_url   TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_learning_lessons_section ON learning_lessons (section_id);

CREATE TABLE IF NOT EXISTS lesson_progress (
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  lesson_id     UUID NOT NULL REFERENCES learning_lessons (id) ON DELETE CASCADE,
  completed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, lesson_id)
);

CREATE INDEX IF NOT EXISTS idx_lesson_progress_lesson ON lesson_progress (lesson_id);
