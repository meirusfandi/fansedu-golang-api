CREATE TABLE IF NOT EXISTS analytics_events (
  id UUID PRIMARY KEY,
  session_id VARCHAR(255) NOT NULL,
  event TEXT NOT NULL,
  page TEXT NOT NULL,
  label TEXT,
  program_id UUID,
  program_slug TEXT,
  metadata JSONB,
  ip_address VARCHAR(64),
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at
  ON analytics_events(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_events_session_event
  ON analytics_events(session_id, event, page, label, created_at DESC);

