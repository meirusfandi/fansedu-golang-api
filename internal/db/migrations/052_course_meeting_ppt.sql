-- Materi PPT per pertemuan (admin/trainer upload → URL disimpan, siswa baca lewat journey).

ALTER TABLE course_meetings
  ADD COLUMN IF NOT EXISTS ppt_url TEXT;

COMMENT ON COLUMN course_meetings.ppt_url IS 'URL file PowerPoint (.ppt/.pptx), bisa dari POST .../upload/course-material';

ALTER TABLE learning_lessons
  ADD COLUMN IF NOT EXISTS ppt_url TEXT;

COMMENT ON COLUMN learning_lessons.ppt_url IS 'Link materi PPT untuk lesson (mirror pdf_url)';
