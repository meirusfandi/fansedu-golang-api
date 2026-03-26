-- Hubungan kelas (course) ↔ tryout_session untuk LMS: satu kelas bisa punya banyak tryout terhubung (urutan sort_order).

CREATE TABLE IF NOT EXISTS course_tryouts (
  course_id          UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  tryout_session_id  UUID NOT NULL REFERENCES tryout_sessions(id) ON DELETE CASCADE,
  sort_order         INT NOT NULL DEFAULT 0,
  PRIMARY KEY (course_id, tryout_session_id)
);
CREATE INDEX IF NOT EXISTS idx_course_tryouts_tryout ON course_tryouts(tryout_session_id);

COMMENT ON TABLE course_tryouts IS 'Tryout yang ditampilkan/dihubungkan ke sebuah kelas LMS';
