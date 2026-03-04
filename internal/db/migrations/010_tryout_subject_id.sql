-- Tryout/event per bidang (subject): siswa hanya melihat tryout yang sesuai subject-nya

ALTER TABLE tryout_sessions
  ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_tryout_sessions_subject ON tryout_sessions (subject_id);
