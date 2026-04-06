-- Add level_id (jenjang pendidikan) ke users profile.
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS level_id UUID REFERENCES levels (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_users_level ON users (level_id);

COMMENT ON COLUMN users.level_id IS 'Jenjang pendidikan user: SD/SMP/SMA (relasi ke levels)';
