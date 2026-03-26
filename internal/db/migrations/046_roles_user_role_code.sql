-- Menjembatani slug publik (tabel roles) dengan nilai kolom users.role (tipe user_role enum).

ALTER TABLE roles ADD COLUMN IF NOT EXISTS user_role_code VARCHAR(64);
COMMENT ON COLUMN roles.user_role_code IS 'Nilai untuk users.role / JWT (label enum user_role PostgreSQL)';

UPDATE roles SET user_role_code = 'student' WHERE slug = 'siswa' AND (user_role_code IS NULL OR trim(user_role_code) = '');
UPDATE roles SET user_role_code = 'guru' WHERE slug IN ('pengajar', 'pembimbing') AND (user_role_code IS NULL OR trim(user_role_code) = '');
UPDATE roles SET user_role_code = 'admin' WHERE slug = 'admin' AND (user_role_code IS NULL OR trim(user_role_code) = '');

-- Slug sama dengan label enum (untuk role admin extended / teknis)
UPDATE roles SET user_role_code = slug
WHERE user_role_code IS NULL OR trim(user_role_code) = ''
  AND slug IN (
    'student', 'guru', 'trainer', 'instructor',
    'super_admin', 'finance_admin', 'academic_admin', 'content_admin'
  );

-- Seed role admin extended jika belum ada baris di roles
INSERT INTO roles (name, slug, description, user_role_code) VALUES
  ('Super Admin', 'super_admin', 'Akses penuh', 'super_admin'),
  ('Finance Admin', 'finance_admin', 'Pembayaran & laporan keuangan', 'finance_admin'),
  ('Academic Admin', 'academic_admin', 'Pengguna, kursus, tryout', 'academic_admin'),
  ('Content Admin', 'content_admin', 'Konten, landing, master data terbatas', 'content_admin'),
  ('Trainer', 'trainer', 'Pelatih', 'trainer')
ON CONFLICT (slug) DO UPDATE SET
  user_role_code = COALESCE(NULLIF(trim(roles.user_role_code), ''), EXCLUDED.user_role_code);

-- Baris lama tanpa mapping: pakai slug jika cocok dengan enum, selain itu fallback student
UPDATE roles SET user_role_code = slug
WHERE (user_role_code IS NULL OR trim(user_role_code) = '')
  AND slug IN ('student', 'guru', 'trainer', 'instructor', 'admin');

UPDATE roles SET user_role_code = 'student'
WHERE user_role_code IS NULL OR trim(user_role_code) = '';
