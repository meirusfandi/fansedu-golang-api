-- Mode penilaian tryout: auto (kunci + sistem) atau manual (admin/guru isi skor per soal).

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tryout_grading_mode') THEN
    CREATE TYPE tryout_grading_mode AS ENUM ('auto', 'manual');
  END IF;
END $$;

ALTER TABLE tryout_sessions
  ADD COLUMN IF NOT EXISTS grading_mode tryout_grading_mode NOT NULL DEFAULT 'auto';
