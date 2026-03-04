-- Seed data: roles, dan data master awal (idempotent)

-- 1. ROLES: Admin, Pembimbing, Siswa, Pengajar
INSERT INTO roles (name, slug, description) VALUES
  ('Admin', 'admin', 'Administrator sistem dengan akses penuh'),
  ('Pembimbing', 'pembimbing', 'Pembimbing siswa dan mentor'),
  ('Siswa', 'siswa', 'Siswa atau peserta belajar'),
  ('Pengajar', 'pengajar', 'Pengajar atau guru')
ON CONFLICT (slug) DO NOTHING;

-- 2. SETTINGS: pengaturan dasar (opsional)
INSERT INTO settings (key, slug, value, description) VALUES
  ('site_name', 'site-name', 'FansEdu LMS', 'Nama situs'),
  ('site_description', 'site-description', 'Platform pembelajaran dan tryout', 'Deskripsi singkat situs'),
  ('maintenance_mode', 'maintenance-mode', 'false', 'Mode maintenance (true/false)')
ON CONFLICT (key) DO NOTHING;

-- 3. SUBJECTS: contoh mata pelajaran (opsional)
INSERT INTO subjects (name, slug, description, sort_order) VALUES
  ('Informatika', 'informatika', 'Mata pelajaran Informatika / Pemrograman', 1),
  ('Matematika', 'matematika', 'Mata pelajaran Matematika', 2),
  ('Fisika', 'fisika', 'Mata pelajaran Fisika', 3)
ON CONFLICT (slug) DO NOTHING;
