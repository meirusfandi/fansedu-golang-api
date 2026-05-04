-- Extend course content types and lesson recording support.
-- New course content types: article, zoom, recording.
-- New lesson/program field: recording_url.
-- Note: ALTER TYPE ADD VALUE IF NOT EXISTS requires PostgreSQL 9.6+
--       and is safe to call outside a transaction block (PG 12+ allows in-transaction).

ALTER TYPE course_content_type ADD VALUE IF NOT EXISTS 'article';
ALTER TYPE course_content_type ADD VALUE IF NOT EXISTS 'zoom';
ALTER TYPE course_content_type ADD VALUE IF NOT EXISTS 'recording';

ALTER TABLE course_meetings
  ADD COLUMN IF NOT EXISTS recording_url TEXT;

ALTER TABLE learning_lessons
  ADD COLUMN IF NOT EXISTS recording_url TEXT;

COMMENT ON COLUMN course_meetings.recording_url IS 'Link rekaman kelas per pertemuan (opsional).';
COMMENT ON COLUMN learning_lessons.recording_url IS 'Link rekaman kelas untuk ditampilkan pada learning journey.';
