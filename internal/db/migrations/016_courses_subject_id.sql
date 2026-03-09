-- Courses per bidang (subject): filter kelas berdasarkan subject siswa
ALTER TABLE courses
  ADD COLUMN IF NOT EXISTS subject_id UUID REFERENCES subjects (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_courses_subject ON courses (subject_id);

COMMENT ON COLUMN courses.subject_id IS 'Bidang/mata pelajaran; NULL = umum (tampil untuk semua siswa)';
