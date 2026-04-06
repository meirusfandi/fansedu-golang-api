-- AI Question Engine (question bank, submissions, subscriptions)

CREATE TABLE IF NOT EXISTS ai_questions (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject            VARCHAR(50) NOT NULL,   -- math | informatics
  grade              VARCHAR(20) NOT NULL,   -- sd | smp | sma
  topic              VARCHAR(100) NOT NULL,  -- dp, graph, aritmatika, dll
  difficulty         VARCHAR(20) NOT NULL,   -- easy | medium | hard | olympiad
  question_type      VARCHAR(20) NOT NULL DEFAULT 'mcq',
  question_text      TEXT NOT NULL,
  choices_json       JSONB NOT NULL DEFAULT '[]'::jsonb,
  correct_answer     TEXT NOT NULL,
  explanation        TEXT NOT NULL DEFAULT '',
  concept_tags       JSONB NOT NULL DEFAULT '[]'::jsonb,
  estimated_time_sec INT NOT NULL DEFAULT 300,
  is_active          BOOLEAN NOT NULL DEFAULT TRUE,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_questions_filter
  ON ai_questions (subject, grade, topic, difficulty, is_active);

DROP TRIGGER IF EXISTS ai_questions_updated_at ON ai_questions;
CREATE TRIGGER ai_questions_updated_at
  BEFORE UPDATE ON ai_questions
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE IF NOT EXISTS ai_submissions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  question_id   UUID NOT NULL REFERENCES ai_questions (id) ON DELETE CASCADE,
  answer        TEXT NOT NULL,
  is_correct    BOOLEAN NOT NULL,
  time_spent_ms BIGINT NOT NULL DEFAULT 0,
  submitted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_submissions_user_time
  ON ai_submissions (user_id, submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_submissions_question
  ON ai_submissions (question_id);

CREATE TABLE IF NOT EXISTS ai_subscriptions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  plan_code  VARCHAR(50) NOT NULL,
  status     VARCHAR(30) NOT NULL DEFAULT 'active',
  start_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  end_at     TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_subscriptions_user_status
  ON ai_subscriptions (user_id, status, end_at);
