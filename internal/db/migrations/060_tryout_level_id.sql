-- Add level_id FK to tryout_sessions so frontend can send levelId
-- and backend resolves school_level from levels table.
ALTER TABLE tryout_sessions
  ADD COLUMN IF NOT EXISTS level_id UUID REFERENCES levels (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_tryout_sessions_level_id ON tryout_sessions (level_id);
