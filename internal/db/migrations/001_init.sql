-- FansEdu LMS - Database Schema (PostgreSQL)
-- Mendukung: auth, admin, siswa, event tryout, soal, attempt, jawaban, rekomendasi, kursus, sertifikat

-- ---------------------------------------------------------------------------
-- 1. USERS & AUTH
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
    CREATE TYPE user_role AS ENUM ('admin', 'student');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email             VARCHAR(255) NOT NULL UNIQUE,
  password_hash     VARCHAR(255) NOT NULL,
  name              VARCHAR(255) NOT NULL,
  role              user_role NOT NULL DEFAULT 'student',
  avatar_url        VARCHAR(512),
  email_verified_at TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

-- Forgot password
CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires ON password_reset_tokens (expires_at);

-- ---------------------------------------------------------------------------
-- 2. TRYOUT SESSIONS (Event / Jadwal simulasi OSN)
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tryout_level') THEN
    CREATE TYPE tryout_level AS ENUM ('easy', 'medium', 'hard');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tryout_status') THEN
    CREATE TYPE tryout_status AS ENUM ('draft', 'open', 'closed');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS tryout_sessions (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title              VARCHAR(500) NOT NULL,
  short_title        VARCHAR(100),
  description        TEXT,
  duration_minutes   INTEGER NOT NULL,
  questions_count    INTEGER NOT NULL,
  level              tryout_level NOT NULL DEFAULT 'medium',
  opens_at           TIMESTAMPTZ NOT NULL,
  closes_at          TIMESTAMPTZ NOT NULL,
  max_participants   INTEGER,
  status             tryout_status NOT NULL DEFAULT 'open',
  created_by         UUID REFERENCES users (id),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tryout_sessions_status ON tryout_sessions (status);
CREATE INDEX IF NOT EXISTS idx_tryout_sessions_opens_closes ON tryout_sessions (opens_at, closes_at);

-- ---------------------------------------------------------------------------
-- 3. QUESTIONS (Soal per tryout)
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'question_type') THEN
    CREATE TYPE question_type AS ENUM ('short', 'multiple_choice', 'true_false');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS questions (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tryout_session_id   UUID NOT NULL REFERENCES tryout_sessions (id) ON DELETE CASCADE,
  sort_order          INTEGER NOT NULL,
  type                question_type NOT NULL,
  body                TEXT NOT NULL,
  options             JSONB,
  max_score           NUMERIC(5,2) NOT NULL DEFAULT 1,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_questions_tryout_session ON questions (tryout_session_id);

-- ---------------------------------------------------------------------------
-- 4. ATTEMPTS (Siswa mengerjakan satu tryout)
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'attempt_status') THEN
    CREATE TYPE attempt_status AS ENUM ('in_progress', 'submitted', 'expired');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS attempts (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id             UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  tryout_session_id   UUID NOT NULL REFERENCES tryout_sessions (id) ON DELETE CASCADE,
  started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  submitted_at        TIMESTAMPTZ,
  status              attempt_status NOT NULL DEFAULT 'in_progress',
  score               NUMERIC(6,2),
  max_score           NUMERIC(6,2),
  percentile          NUMERIC(5,2),
  time_seconds_spent  INTEGER,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, tryout_session_id)
);

CREATE INDEX IF NOT EXISTS idx_attempts_user ON attempts (user_id);
CREATE INDEX IF NOT EXISTS idx_attempts_tryout_session ON attempts (tryout_session_id);
CREATE INDEX IF NOT EXISTS idx_attempts_status ON attempts (status);

-- ---------------------------------------------------------------------------
-- 5. ATTEMPT ANSWERS (Jawaban per soal per attempt)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS attempt_answers (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  attempt_id       UUID NOT NULL REFERENCES attempts (id) ON DELETE CASCADE,
  question_id      UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
  answer_text      TEXT,
  selected_option  VARCHAR(50),
  is_marked        BOOLEAN NOT NULL DEFAULT FALSE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (attempt_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_attempt_answers_attempt ON attempt_answers (attempt_id);

-- ---------------------------------------------------------------------------
-- 6. ATTEMPT FEEDBACK (Ringkasan & rangkuman per attempt - bisa dari AI)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS attempt_feedback (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  attempt_id         UUID NOT NULL REFERENCES attempts (id) ON DELETE CASCADE UNIQUE,
  summary            TEXT,
  recap              TEXT,
  strength_areas     JSONB,
  improvement_areas  JSONB,
  recommendation_text TEXT,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- 7. COURSES / KELAS PEMBINAAN
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS courses (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title       VARCHAR(500) NOT NULL,
  description TEXT,
  created_by  UUID REFERENCES users (id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- 8. COURSE ENROLLMENTS
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'enrollment_status') THEN
    CREATE TYPE enrollment_status AS ENUM ('enrolled', 'in_progress', 'completed');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS course_enrollments (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  course_id   UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  status      enrollment_status NOT NULL DEFAULT 'enrolled',
  enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, course_id)
);

CREATE INDEX IF NOT EXISTS idx_course_enrollments_user ON course_enrollments (user_id);
CREATE INDEX IF NOT EXISTS idx_course_enrollments_course ON course_enrollments (course_id);

-- ---------------------------------------------------------------------------
-- 9. CERTIFICATES
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS certificates (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id             UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  tryout_session_id   UUID REFERENCES tryout_sessions (id) ON DELETE SET NULL,
  course_id           UUID REFERENCES courses (id) ON DELETE SET NULL,
  issued_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (
    (tryout_session_id IS NOT NULL AND course_id IS NULL) OR
    (tryout_session_id IS NULL AND course_id IS NOT NULL)
  )
);

CREATE INDEX IF NOT EXISTS idx_certificates_user ON certificates (user_id);

-- ---------------------------------------------------------------------------
-- Trigger: updated_at
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
DROP TRIGGER IF EXISTS tryout_sessions_updated_at ON tryout_sessions;
CREATE TRIGGER tryout_sessions_updated_at BEFORE UPDATE ON tryout_sessions
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
DROP TRIGGER IF EXISTS attempts_updated_at ON attempts;
CREATE TRIGGER attempts_updated_at BEFORE UPDATE ON attempts
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
DROP TRIGGER IF EXISTS attempt_answers_updated_at ON attempt_answers;
CREATE TRIGGER attempt_answers_updated_at BEFORE UPDATE ON attempt_answers
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
DROP TRIGGER IF EXISTS attempt_feedback_updated_at ON attempt_feedback;
CREATE TRIGGER attempt_feedback_updated_at BEFORE UPDATE ON attempt_feedback
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
DROP TRIGGER IF EXISTS courses_updated_at ON courses;
CREATE TRIGGER courses_updated_at BEFORE UPDATE ON courses
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
