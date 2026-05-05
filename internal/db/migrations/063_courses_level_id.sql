-- Jenjang (SD/SMP/SMA) opsional per kelas; filter/query sama seperti tryout/users.
ALTER TABLE courses
  ADD COLUMN IF NOT EXISTS level_id UUID REFERENCES levels (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_courses_level_id ON courses (level_id);

COMMENT ON COLUMN courses.level_id IS 'Jenjang pendidikan (levels.id); NULL = tidak ditautkan';
