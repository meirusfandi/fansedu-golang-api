-- Guru + trainer roles; paid slots + trainer–student linking
-- guru = pendaftaran mandiri; trainer = nanti dibuat admin

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    WHERE t.typname = 'user_role' AND e.enumlabel = 'guru'
  ) THEN
    ALTER TYPE user_role ADD VALUE 'guru';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    WHERE t.typname = 'user_role' AND e.enumlabel = 'trainer'
  ) THEN
    ALTER TYPE user_role ADD VALUE 'trainer';
  END IF;
END $$;

-- One row per trainer: paid_slots = number of slots purchased (each slot = one student allowed)
CREATE TABLE IF NOT EXISTS trainer_slots (
  trainer_id   UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  paid_slots   INTEGER NOT NULL DEFAULT 0 CHECK (paid_slots >= 0),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trainer_slots_trainer ON trainer_slots (trainer_id);

-- Students registered by a trainer (count must not exceed trainer_slots.paid_slots)
CREATE TABLE IF NOT EXISTS trainer_students (
  trainer_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  student_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (trainer_id, student_id),
  CHECK (trainer_id != student_id)
);

CREATE INDEX IF NOT EXISTS idx_trainer_students_trainer ON trainer_students (trainer_id);
CREATE INDEX IF NOT EXISTS idx_trainer_students_student ON trainer_students (student_id);

DROP TRIGGER IF EXISTS trainer_slots_updated_at ON trainer_slots;
CREATE TRIGGER trainer_slots_updated_at BEFORE UPDATE ON trainer_slots
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
