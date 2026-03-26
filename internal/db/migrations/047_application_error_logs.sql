-- Log error HTTP & panic untuk analitik / triase admin (semua role pengguna).

CREATE TABLE IF NOT EXISTS application_error_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  error_type      VARCHAR(48) NOT NULL,
  error_code      VARCHAR(128),
  message         TEXT NOT NULL DEFAULT '',
  http_status     INT NOT NULL,
  method          VARCHAR(16) NOT NULL DEFAULT '',
  path            TEXT NOT NULL,
  query_string    TEXT,
  user_id         UUID REFERENCES users (id) ON DELETE SET NULL,
  user_role       VARCHAR(64),
  request_id      VARCHAR(128),
  ip_address      VARCHAR(128),
  user_agent      TEXT,
  stack_trace     TEXT,
  meta            JSONB NOT NULL DEFAULT '{}',
  resolved_at     TIMESTAMPTZ,
  resolved_by     UUID REFERENCES users (id) ON DELETE SET NULL,
  admin_note      TEXT
);

CREATE INDEX IF NOT EXISTS idx_application_error_logs_created_at
  ON application_error_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_application_error_logs_type_created
  ON application_error_logs (error_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_application_error_logs_status_created
  ON application_error_logs (http_status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_application_error_logs_unresolved
  ON application_error_logs (created_at DESC)
  WHERE resolved_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_application_error_logs_user_id
  ON application_error_logs (user_id)
  WHERE user_id IS NOT NULL;

COMMENT ON TABLE application_error_logs IS 'Error & respons 4xx/5xx + panic; dilihat admin via API';
