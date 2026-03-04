-- Levels (jenjang pendidikan: SD, SMP, SMA) dan relasi subject-level

CREATE TABLE IF NOT EXISTS levels (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  slug        VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  icon_url    VARCHAR(512),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_levels_slug ON levels (slug);
CREATE INDEX IF NOT EXISTS idx_levels_sort ON levels (sort_order);
DROP TRIGGER IF EXISTS levels_updated_at ON levels;
CREATE TRIGGER levels_updated_at BEFORE UPDATE ON levels
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- Relasi many-to-many: mata pelajaran per jenjang (bidang SD, SMP, SMA)
CREATE TABLE IF NOT EXISTS subject_levels (
  subject_id  UUID NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
  level_id    UUID NOT NULL REFERENCES levels (id) ON DELETE CASCADE,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (subject_id, level_id)
);

CREATE INDEX IF NOT EXISTS idx_subject_levels_level ON subject_levels (level_id);
CREATE INDEX IF NOT EXISTS idx_subject_levels_subject ON subject_levels (subject_id);
