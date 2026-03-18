-- =============================================================================
-- FANSEDU LMS — Production PostgreSQL Schema
-- =============================================================================
-- Roles: STUDENT, TEACHER, TRAINER/TUTOR, ADMIN
-- Scale: 100k+ students, 10k+ classes, millions of quiz answers
-- UUID PKs, FKs, indexes, timestamps. SQLC/GORM compatible.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 1. ROLES
-- =============================================================================
CREATE TABLE roles (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  slug        VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE roles IS 'System roles: student, teacher, trainer, admin';
CREATE INDEX idx_roles_slug ON roles (slug);

-- =============================================================================
-- 2. SCHOOLS
-- =============================================================================
CREATE TABLE schools (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           VARCHAR(255) NOT NULL,
  slug           VARCHAR(255) NOT NULL UNIQUE,
  npsn           VARCHAR(50),
  kabupaten_kota VARCHAR(255),
  address        TEXT,
  telepon        VARCHAR(50),
  logo_url       VARCHAR(512),
  description    TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE schools IS 'Schools/institutions; teachers manage profile and students under school';
CREATE INDEX idx_schools_slug ON schools (slug);

-- =============================================================================
-- 3. USERS
-- =============================================================================
CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email             VARCHAR(255) NOT NULL UNIQUE,
  password_hash     VARCHAR(255) NOT NULL,
  name              VARCHAR(255) NOT NULL,
  role_id           UUID NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
  school_id         UUID REFERENCES schools (id) ON DELETE SET NULL,
  avatar_url        VARCHAR(512),
  email_verified_at TIMESTAMPTZ,
  last_login_at     TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE users IS 'All users: students, teachers, trainers, admins';
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role_id ON users (role_id);
CREATE INDEX idx_users_school_id ON users (school_id);
CREATE INDEX idx_users_created_at ON users (created_at);

CREATE TABLE password_reset_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens (user_id);
CREATE INDEX idx_password_reset_tokens_expires ON password_reset_tokens (expires_at);

-- =============================================================================
-- 4. TEACHER–SCHOOL & TEACHER–STUDENT
-- =============================================================================
CREATE TABLE teacher_schools (
  teacher_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  school_id   UUID NOT NULL REFERENCES schools (id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (teacher_id, school_id)
);
COMMENT ON TABLE teacher_schools IS 'Teachers can belong to multiple schools';
CREATE INDEX idx_teacher_schools_school ON teacher_schools (school_id);

CREATE TABLE teacher_students (
  teacher_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  student_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (teacher_id, student_id),
  CHECK (teacher_id != student_id)
);
COMMENT ON TABLE teacher_students IS 'Students under a teacher (for teacher dashboard, enroll into classes)';
CREATE INDEX idx_teacher_students_student ON teacher_students (student_id);

-- Trainer paid slots (trainer buys capacity to add students)
CREATE TABLE trainer_slots (
  trainer_id   UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  paid_slots   INTEGER NOT NULL DEFAULT 0 CHECK (paid_slots >= 0),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE trainer_slots IS 'Paid slots per trainer; each slot allows adding one student';

-- =============================================================================
-- 5. TOPICS (for question bank & weakness analytics)
-- =============================================================================
-- Optional: subjects table for grouping topics (e.g. Mathematics -> algebra, geometry)
CREATE TABLE subjects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(255) NOT NULL,
  slug        VARCHAR(255) NOT NULL UNIQUE,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_subjects_slug ON subjects (slug);

CREATE TABLE topics (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id  UUID REFERENCES subjects (id) ON DELETE SET NULL,
  name        VARCHAR(255) NOT NULL,
  slug        VARCHAR(255) NOT NULL UNIQUE,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE topics IS 'Topics (e.g. algebra, geometry) for question bank and error_rate analytics';
CREATE INDEX idx_topics_slug ON topics (slug);
CREATE INDEX idx_topics_subject ON topics (subject_id);

-- =============================================================================
-- 6. CLASSES (created by trainer)
-- =============================================================================
CREATE TYPE class_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE classes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  trainer_id  UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
  title       VARCHAR(500) NOT NULL,
  description TEXT,
  capacity    INTEGER NOT NULL DEFAULT 50 CHECK (capacity > 0),
  price INTEGER NOT NULL DEFAULT 0,
  status      class_status NOT NULL DEFAULT 'draft',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE classes IS 'Classes created by trainer; capacity = max students';
CREATE INDEX idx_classes_trainer ON classes (trainer_id);
CREATE INDEX idx_classes_status ON classes (status);
CREATE INDEX idx_classes_created_at ON classes (created_at);

-- =============================================================================
-- 7. CLASS ENROLLMENTS
-- =============================================================================
CREATE TYPE enrollment_status AS ENUM ('enrolled', 'in_progress', 'completed', 'dropped');

CREATE TABLE class_enrollments (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id     UUID NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
  student_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  status       enrollment_status NOT NULL DEFAULT 'enrolled',
  enrolled_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (class_id, student_id)
);
COMMENT ON TABLE class_enrollments IS 'Student enrollments in classes; used for access to modules/quizzes and ranking';
CREATE INDEX idx_class_enrollments_class ON class_enrollments (class_id);
CREATE INDEX idx_class_enrollments_student ON class_enrollments (student_id);
CREATE INDEX idx_class_enrollments_status ON class_enrollments (status);

-- =============================================================================
-- 8. MODULES (learning content per class)
-- =============================================================================
CREATE TYPE module_content_type AS ENUM ('video', 'document', 'link', 'embed', 'text');

CREATE TABLE modules (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id     UUID NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
  title        VARCHAR(500) NOT NULL,
  description  TEXT,
  sort_order   INTEGER NOT NULL DEFAULT 0,
  content_type module_content_type NOT NULL DEFAULT 'text',
  content_url  TEXT,
  content_body TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE modules IS 'Learning modules per class; can contain materials and quizzes';
CREATE INDEX idx_modules_class ON modules (class_id);
CREATE INDEX idx_modules_sort ON modules (class_id, sort_order);

-- =============================================================================
-- 9. QUIZZES (per module or per class)
-- =============================================================================
CREATE TABLE quizzes (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id           UUID NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
  module_id          UUID REFERENCES modules (id) ON DELETE SET NULL,
  title              VARCHAR(500) NOT NULL,
  description        TEXT,
  passing_score_pct  NUMERIC(5,2) DEFAULT 60,
  time_limit_minutes INTEGER,
  sort_order         INTEGER NOT NULL DEFAULT 0,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE quizzes IS 'Quizzes in a class; can be attached to a module or standalone';
CREATE INDEX idx_quizzes_class ON quizzes (class_id);
CREATE INDEX idx_quizzes_module ON quizzes (module_id);

-- =============================================================================
-- 10. QUESTIONS (question bank + tryout questions)
-- =============================================================================
CREATE TYPE question_type AS ENUM ('short', 'multiple_choice', 'true_false', 'essay');

CREATE TABLE questions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic_id      UUID REFERENCES topics (id) ON DELETE SET NULL,
  type          question_type NOT NULL,
  body          TEXT NOT NULL,
  body_html     TEXT,
  max_score     NUMERIC(5,2) NOT NULL DEFAULT 1,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE questions IS 'Question bank; topic_id for weakness analysis; used in quizzes and tryouts';
CREATE INDEX idx_questions_topic ON questions (topic_id);
CREATE INDEX idx_questions_type ON questions (type);
CREATE INDEX idx_questions_created_at ON questions (created_at);

CREATE TABLE question_options (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
  option_key  VARCHAR(10) NOT NULL,
  option_text TEXT NOT NULL,
  is_correct  BOOLEAN NOT NULL DEFAULT FALSE,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (question_id, option_key)
);
COMMENT ON TABLE question_options IS 'Multiple choice options per question';
CREATE INDEX idx_question_options_question ON question_options (question_id);

CREATE TABLE quiz_questions (
  quiz_id     UUID NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
  question_id UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (quiz_id, question_id)
);
COMMENT ON TABLE quiz_questions IS 'Questions in a quiz; enables generate quiz from question bank';
CREATE INDEX idx_quiz_questions_question ON quiz_questions (question_id);

-- =============================================================================
-- 11. QUIZ ATTEMPTS & STUDENT ANSWERS (class quizzes)
-- =============================================================================
CREATE TYPE attempt_status AS ENUM ('in_progress', 'submitted', 'expired');

CREATE TABLE quiz_attempts (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  quiz_id        UUID NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
  score          NUMERIC(10,2),
  max_score      NUMERIC(10,2),
  status         attempt_status NOT NULL DEFAULT 'in_progress',
  started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  submitted_at   TIMESTAMPTZ,
  time_seconds   INTEGER,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE quiz_attempts IS 'One attempt per student per quiz; source for student_scores and ranking';
CREATE INDEX idx_quiz_attempts_student ON quiz_attempts (student_id);
CREATE INDEX idx_quiz_attempts_quiz ON quiz_attempts (quiz_id);
CREATE INDEX idx_quiz_attempts_status ON quiz_attempts (status);
CREATE INDEX idx_quiz_attempts_submitted ON quiz_attempts (submitted_at) WHERE submitted_at IS NOT NULL;

CREATE TABLE student_answers (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  attempt_id       UUID NOT NULL REFERENCES quiz_attempts (id) ON DELETE CASCADE,
  question_id      UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
  question_option_id UUID REFERENCES question_options (id) ON DELETE SET NULL,
  answer_text      TEXT,
  is_correct       BOOLEAN,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (attempt_id, question_id)
);
COMMENT ON TABLE student_answers IS 'Per-question answers in a quiz attempt; used for weakness analysis';
CREATE INDEX idx_student_answers_attempt ON student_answers (attempt_id);
CREATE INDEX idx_student_answers_question ON student_answers (question_id);
CREATE INDEX idx_student_answers_is_correct ON student_answers (question_id, is_correct);

-- =============================================================================
-- 12. TRYOUTS & REGISTRATIONS
-- =============================================================================
CREATE TYPE tryout_status AS ENUM ('draft', 'open', 'closed');

CREATE TABLE tryouts (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title              VARCHAR(500) NOT NULL,
  description        TEXT,
  duration_minutes   INTEGER NOT NULL,
  opens_at           TIMESTAMPTZ NOT NULL,
  closes_at          TIMESTAMPTZ NOT NULL,
  status             tryout_status NOT NULL DEFAULT 'open',
  created_by         UUID REFERENCES users (id) ON DELETE SET NULL,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE tryouts IS 'Tryout events; students register and take attempts';
CREATE INDEX idx_tryouts_status ON tryouts (status);
CREATE INDEX idx_tryouts_opens_closes ON tryouts (opens_at, closes_at);

CREATE TABLE tryout_registrations (
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  tryout_id     UUID NOT NULL REFERENCES tryouts (id) ON DELETE CASCADE,
  registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, tryout_id)
);
CREATE INDEX idx_tryout_registrations_tryout ON tryout_registrations (tryout_id);

-- Tryout questions (which questions belong to a tryout)
CREATE TABLE tryout_questions (
  tryout_id    UUID NOT NULL REFERENCES tryouts (id) ON DELETE CASCADE,
  question_id  UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
  sort_order   INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (tryout_id, question_id)
);
CREATE INDEX idx_tryout_questions_question ON tryout_questions (question_id);

-- Tryout attempts (separate from quiz_attempts for tryout leaderboard)
CREATE TABLE tryout_attempts (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  tryout_id        UUID NOT NULL REFERENCES tryouts (id) ON DELETE CASCADE,
  score            NUMERIC(10,2),
  max_score        NUMERIC(10,2),
  status           attempt_status NOT NULL DEFAULT 'in_progress',
  started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  submitted_at     TIMESTAMPTZ,
  time_seconds     INTEGER,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, tryout_id)
);
CREATE INDEX idx_tryout_attempts_user ON tryout_attempts (user_id);
CREATE INDEX idx_tryout_attempts_tryout ON tryout_attempts (tryout_id);

CREATE TABLE tryout_answer (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  attempt_id       UUID NOT NULL REFERENCES tryout_attempts (id) ON DELETE CASCADE,
  question_id      UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
  question_option_id UUID REFERENCES question_options (id) ON DELETE SET NULL,
  answer_text      TEXT,
  is_correct       BOOLEAN,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (attempt_id, question_id)
);
CREATE INDEX idx_tryout_answer_attempt ON tryout_answer (attempt_id);

-- =============================================================================
-- 13. ORDERS & PAYMENTS
-- =============================================================================
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');
CREATE TYPE order_item_type AS ENUM ('class', 'tryout', 'teacher_slots', 'other');
CREATE TYPE payment_method AS ENUM ('bank_transfer', 'e_wallet', 'other');
CREATE TYPE payment_confirmation_status AS ENUM ('pending', 'confirmed', 'rejected');

CREATE TABLE orders (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  total_price   INTEGER NOT NULL DEFAULT 0,
  currency      VARCHAR(3) NOT NULL DEFAULT 'IDR',
  status        order_status NOT NULL DEFAULT 'pending',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE orders IS 'Order header; checkout creates order + order_items';
CREATE INDEX idx_orders_user ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_created_at ON orders (created_at);

CREATE TABLE order_items (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id     UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
  item_type    order_item_type NOT NULL,
  item_id      UUID,
  quantity     INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
  unit_price   INTEGER NOT NULL,
  total_price  INTEGER NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE order_items IS 'Line items: class purchase, tryout, teacher slots, etc.';
CREATE INDEX idx_order_items_order ON order_items (order_id);
CREATE INDEX idx_order_items_item ON order_items (item_type, item_id);

CREATE TABLE payments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
  amount          INTEGER NOT NULL,
  method          payment_method NOT NULL DEFAULT 'bank_transfer',
  status          payment_confirmation_status NOT NULL DEFAULT 'pending',
  proof_url       VARCHAR(1024),
  confirmed_by    UUID REFERENCES users (id) ON DELETE SET NULL,
  confirmed_at    TIMESTAMPTZ,
  rejection_note  TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE payments IS 'Payment records; proof_url for bank transfer; admin confirms';
CREATE INDEX idx_payments_order ON payments (order_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_created_at ON payments (created_at);

-- =============================================================================
-- 14. NOTIFICATIONS
-- =============================================================================
CREATE TABLE notifications (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  title      VARCHAR(500) NOT NULL,
  body       TEXT,
  type       VARCHAR(50) DEFAULT 'system',
  read_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE notifications IS 'User notifications: payment confirmed, new module, etc.';
CREATE INDEX idx_notifications_user ON notifications (user_id);
CREATE INDEX idx_notifications_read_at ON notifications (user_id, read_at);
CREATE INDEX idx_notifications_created_at ON notifications (created_at);

-- =============================================================================
-- 15. COMMUNICATION: CLASS CHAT & DISCUSSION
-- =============================================================================
CREATE TABLE class_messages (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id  UUID NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
  user_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  message   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE class_messages IS 'Class chat messages';
CREATE INDEX idx_class_messages_class ON class_messages (class_id);
CREATE INDEX idx_class_messages_created_at ON class_messages (class_id, created_at);

CREATE TABLE discussions (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id  UUID NOT NULL REFERENCES classes (id) ON DELETE CASCADE,
  user_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  title     VARCHAR(500) NOT NULL,
  body      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE discussions IS 'Discussion forum threads per class';
CREATE INDEX idx_discussions_class ON discussions (class_id);
CREATE INDEX idx_discussions_created_at ON discussions (class_id, created_at);

CREATE TABLE discussion_replies (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  discussion_id UUID NOT NULL REFERENCES discussions (id) ON DELETE CASCADE,
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  body          TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_discussion_replies_discussion ON discussion_replies (discussion_id);
CREATE INDEX idx_discussion_replies_created_at ON discussion_replies (discussion_id, created_at);

-- =============================================================================
-- 16. STUDENT SCORES (materialized or view from quiz_attempts)
-- =============================================================================
-- Table for denormalized/cached scores if needed; else use view below
CREATE TABLE student_scores (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  quiz_id      UUID NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
  attempt_id   UUID NOT NULL REFERENCES quiz_attempts (id) ON DELETE CASCADE,
  score        NUMERIC(10,2) NOT NULL,
  max_score    NUMERIC(10,2) NOT NULL,
  submitted_at TIMESTAMPTZ NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (attempt_id)
);
COMMENT ON TABLE student_scores IS 'One row per submitted quiz attempt; can be populated by trigger or app';
CREATE INDEX idx_student_scores_student ON student_scores (student_id);
CREATE INDEX idx_student_scores_quiz ON student_scores (quiz_id);
CREATE INDEX idx_student_scores_submitted ON student_scores (submitted_at);

-- =============================================================================
-- TRIGGERS: updated_at
-- =============================================================================
CREATE TRIGGER roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER schools_updated_at BEFORE UPDATE ON schools FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER trainer_slots_updated_at BEFORE UPDATE ON trainer_slots FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER topics_updated_at BEFORE UPDATE ON topics FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER subjects_updated_at BEFORE UPDATE ON subjects FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER classes_updated_at BEFORE UPDATE ON classes FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER class_enrollments_updated_at BEFORE UPDATE ON class_enrollments FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER modules_updated_at BEFORE UPDATE ON modules FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER quizzes_updated_at BEFORE UPDATE ON quizzes FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER questions_updated_at BEFORE UPDATE ON questions FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER quiz_attempts_updated_at BEFORE UPDATE ON quiz_attempts FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER tryouts_updated_at BEFORE UPDATE ON tryouts FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER tryout_attempts_updated_at BEFORE UPDATE ON tryout_attempts FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER discussions_updated_at BEFORE UPDATE ON discussions FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
CREATE TRIGGER discussion_replies_updated_at BEFORE UPDATE ON discussion_replies FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- =============================================================================
-- VIEWS: Class ranking & Student weakness analysis
-- =============================================================================

-- Class ranking: per class, student rank by total quiz score (submitted attempts only)
CREATE OR REPLACE VIEW class_ranking AS
SELECT
  c.id AS class_id,
  c.title AS class_title,
  ce.student_id,
  u.name AS student_name,
  u.email AS student_email,
  COALESCE(SUM(qa.score), 0) AS total_score,
  COALESCE(SUM(qa.max_score), 0) AS total_max_score,
  RANK() OVER (PARTITION BY c.id ORDER BY COALESCE(SUM(qa.score), 0) DESC, MAX(qa.submitted_at) ASC) AS rank_in_class
FROM classes c
JOIN class_enrollments ce ON ce.class_id = c.id AND ce.status IN ('enrolled', 'in_progress', 'completed')
JOIN users u ON u.id = ce.student_id
LEFT JOIN quizzes q ON q.class_id = c.id
LEFT JOIN quiz_attempts qa ON qa.quiz_id = q.id AND qa.student_id = ce.student_id AND qa.status = 'submitted' AND qa.submitted_at IS NOT NULL
GROUP BY c.id, c.title, ce.student_id, u.name, u.email;

COMMENT ON VIEW class_ranking IS 'Per-class student ranking by total quiz score (for leaderboard)';

-- Student weakness analysis: error_rate per topic based on answered questions
CREATE OR REPLACE VIEW student_weakness_analysis AS
SELECT
  qa.student_id,
  q.topic_id,
  t.name AS topic_name,
  COUNT(*) AS total_answered,
  COUNT(*) FILTER (WHERE sa.is_correct = TRUE) AS correct_count,
  COUNT(*) FILTER (WHERE sa.is_correct = FALSE) AS wrong_count,
  ROUND(
    (COUNT(*) FILTER (WHERE sa.is_correct = FALSE)::NUMERIC / NULLIF(COUNT(*), 0)) * 100,
    2
  ) AS error_rate_pct
FROM quiz_attempts qa
JOIN student_answers sa ON sa.attempt_id = qa.id
JOIN questions q ON q.id = sa.question_id
LEFT JOIN topics t ON t.id = q.topic_id
WHERE qa.status = 'submitted' AND qa.submitted_at IS NOT NULL
  AND sa.is_correct IS NOT NULL
GROUP BY qa.student_id, q.topic_id, t.name;

COMMENT ON VIEW student_weakness_analysis IS 'Per-student per-topic error rate for weakness analytics';

-- Optional: view for student score per quiz (simple)
CREATE OR REPLACE VIEW student_quiz_scores AS
SELECT
  qa.student_id,
  qa.quiz_id,
  q.title AS quiz_title,
  q.class_id,
  qa.score,
  qa.max_score,
  qa.submitted_at,
  CASE WHEN qa.max_score > 0 THEN ROUND((qa.score / qa.max_score) * 100, 2) ELSE NULL END AS score_pct
FROM quiz_attempts qa
JOIN quizzes q ON q.id = qa.quiz_id
WHERE qa.status = 'submitted' AND qa.submitted_at IS NOT NULL;

COMMENT ON VIEW student_quiz_scores IS 'Student score per quiz for analytics';

-- =============================================================================
-- SEED: Default roles (run once)
-- =============================================================================
INSERT INTO roles (name, slug, description) VALUES
  ('Student', 'student', 'Siswa yang belajar dan ikut tryout/kelas'),
  ('Teacher', 'teacher', 'Guru yang mengelola sekolah dan siswa'),
  ('Trainer', 'trainer', 'Tutor yang membuat kelas, modul, dan kuis'),
  ('Admin', 'admin', 'Administrator sistem')
ON CONFLICT (slug) DO NOTHING;
