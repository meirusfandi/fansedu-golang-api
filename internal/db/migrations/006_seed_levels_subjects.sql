-- Seed: jenjang pendidikan (SD, SMP, SMA) dan bidang/mata pelajaran per jenjang

-- 1. LEVELS (jenjang pendidikan)
INSERT INTO levels (name, slug, description, sort_order) VALUES
  ('SD', 'sd', 'Sekolah Dasar', 1),
  ('SMP', 'smp', 'Sekolah Menengah Pertama', 2),
  ('SMA', 'sma', 'Sekolah Menengah Atas', 3)
ON CONFLICT (slug) DO NOTHING;

-- 2. SUBJECTS (semua mata pelajaran / bidang) — yang belum ada di 004
INSERT INTO subjects (name, slug, description, sort_order) VALUES
  ('Matematika', 'matematika', 'Matematika', 1),
  ('IPA', 'ipa', 'Ilmu Pengetahuan Alam', 2),
  ('IPS', 'ips', 'Ilmu Pengetahuan Sosial', 3),
  ('Informatika', 'informatika', 'Informatika / Pemrograman', 4),
  ('Biologi', 'biologi', 'Biologi', 5),
  ('Fisika', 'fisika', 'Fisika', 6),
  ('Kimia', 'kimia', 'Kimia', 7),
  ('Geografi', 'geografi', 'Geografi', 8),
  ('Astronomi', 'astronomi', 'Astronomi', 9),
  ('Kebumian', 'kebumian', 'Kebumian', 10),
  ('Ekonomi', 'ekonomi', 'Ekonomi', 11)
ON CONFLICT (slug) DO NOTHING;

-- 3. SUBJECT_LEVELS: bidang per jenjang
-- SD & SMP: Matematika, IPA, IPS
INSERT INTO subject_levels (subject_id, level_id, sort_order)
SELECT s.id, l.id, row_number() OVER (ORDER BY s.sort_order)::int
FROM subjects s
CROSS JOIN levels l
WHERE l.slug IN ('sd', 'smp') AND s.slug IN ('matematika', 'ipa', 'ips')
ON CONFLICT (subject_id, level_id) DO NOTHING;

-- SMA: Matematika, Informatika, Biologi, Fisika, Kimia, Geografi, Astronomi, Kebumian, Ekonomi
INSERT INTO subject_levels (subject_id, level_id, sort_order)
SELECT s.id, l.id, row_number() OVER (ORDER BY s.sort_order)::int
FROM subjects s
CROSS JOIN levels l
WHERE l.slug = 'sma'
  AND s.slug IN ('matematika', 'informatika', 'biologi', 'fisika', 'kimia', 'geografi', 'astronomi', 'kebumian', 'ekonomi')
ON CONFLICT (subject_id, level_id) DO NOTHING;
