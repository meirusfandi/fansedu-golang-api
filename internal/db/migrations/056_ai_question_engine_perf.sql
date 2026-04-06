-- Performance tuning indexes for AI question engine.

CREATE INDEX IF NOT EXISTS idx_ai_questions_updated_active
  ON ai_questions (is_active, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_submissions_user_question
  ON ai_submissions (user_id, question_id, submitted_at DESC);
