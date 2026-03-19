CREATE TABLE IF NOT EXISTS analytics_pageviews (
  id UUID PRIMARY KEY,
  session_id VARCHAR(255) NOT NULL,
  path TEXT NOT NULL,
  ip_address VARCHAR(64),
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_pageviews_created_at
  ON analytics_pageviews(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_pageviews_session_path_created
  ON analytics_pageviews(session_id, path, created_at DESC);
