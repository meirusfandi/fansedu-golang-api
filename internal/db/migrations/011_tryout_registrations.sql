-- Pendaftaran tryout: siswa mendaftar dulu, masuk leaderboard (urut: nama / nilai tertinggi / waktu tercepat / nama)

CREATE TABLE IF NOT EXISTS tryout_registrations (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id             UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  tryout_session_id   UUID NOT NULL REFERENCES tryout_sessions (id) ON DELETE CASCADE,
  registered_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, tryout_session_id)
);

CREATE INDEX IF NOT EXISTS idx_tryout_registrations_tryout ON tryout_registrations (tryout_session_id);
CREATE INDEX IF NOT EXISTS idx_tryout_registrations_user ON tryout_registrations (user_id);

-- Semua siswa otomatis terdaftar untuk tryout yang akan datang (status open/closed, bukan draft)
INSERT INTO tryout_registrations (user_id, tryout_session_id)
SELECT u.id, t.id
FROM users u
CROSS JOIN tryout_sessions t
WHERE u.role = 'student'
  AND t.status != 'draft'
ON CONFLICT (user_id, tryout_session_id) DO NOTHING;
