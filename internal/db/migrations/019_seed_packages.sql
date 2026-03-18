-- =============================================================================
-- Seed: Packages (Program yang Sedang Dibuka)
-- =============================================================================
-- Jalankan setelah landing_schema.sql. Menghapus data lama lalu mengisi 3 paket.
-- psql $DATABASE_URL -f database/seed_packages.sql
-- =============================================================================

DELETE FROM packages;

INSERT INTO packages (
  id, name, slug, short_description,
  cta_label, wa_message_template, cta_url,
  is_open, is_bundle, bundle_subtitle, durasi,
  materi, fasilitas, bonus
) VALUES
-- 1. Algorithm & Programming Foundation
(
  gen_random_uuid(),
  'Algorithm & Programming Foundation',
  'algorithm-programming-foundation',
  'Kelas dasar untuk membangun fondasi berpikir algoritmik dan pemrograman yang dibutuhkan dalam kompetisi informatika.',
  'Lihat Detail / Daftar',
  NULL,
  NULL,
  true,
  false,
  NULL,
  '4 Minggu',
  '["Menyelesaikan soal algoritma dasar","Menggunakan C++ untuk kompetisi","Teknik problem solving olimpiade","Struktur data dasar"]'::jsonb,
  '["2x Live Class per minggu","Latihan soal terstruktur","Rekaman kelas (record class)","Forum diskusi peserta"]'::jsonb,
  '[]'::jsonb
),
-- 2. Pelatihan Intensif OSN-K 2026
(
  gen_random_uuid(),
  'Pelatihan Intensif OSN-K 2026 Informatika',
  'pelatihan-intensif-osn-k-2026',
  'Program pelatihan khusus untuk membantu siswa mempersiapkan seleksi Olimpiade Sains Nasional bidang Informatika.',
  'Lihat Detail / Daftar',
  NULL,
  NULL,
  true,
  false,
  NULL,
  '4 Minggu',
  '["Strategi lolos seleksi OSN tingkat sekolah & kabupaten","Algoritma yang sering keluar di OSN","Soal tipe olimpiade dengan pembahasan mendalam","Problem solving & computational thinking terarah"]'::jsonb,
  '["2x Live Class per minggu","2x Tryout Nasional","Video pembahasan soal","Dashboard ranking nasional peserta"]'::jsonb,
  '[]'::jsonb
),
-- 3. Paket Hemat
(
  gen_random_uuid(),
  'Paket Hemat',
  'paket-hemat-foundation-osn',
  'Foundation + OSN Training — Dapatkan kedua program sekaligus: fondasi algoritma & pemrograman plus persiapan intensif OSN-K. Lebih hemat daripada daftar terpisah.',
  'Lihat Detail / Daftar',
  NULL,
  NULL,
  true,
  true,
  'Foundation + OSN Training',
  '6 Minggu',
  '["Semua keahlian Foundation + OSN-K dalam satu paket","Dari dasar C++ sampai siap menghadapi OSN-K","Akses penuh ke latihan, tryout, dan pembahasan","Lebih hemat, lebih lengkap"]'::jsonb,
  '["2x Live Class per minggu (gabungan kedua program)","Latihan soal terstruktur + 2x Tryout Nasional","Rekaman kelas + video pembahasan soal","Forum diskusi & dashboard ranking nasional"]'::jsonb,
  '["Bank soal OSN","Rekaman kelas","Grup diskusi"]'::jsonb
);

-- Set harga jika kolom sudah ada (026 menambahkan kolom price_early_bird, price_normal)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'packages' AND column_name = 'price_early_bird') THEN
    UPDATE packages SET price_early_bird = 249000, price_normal = 399000 WHERE slug = 'algorithm-programming-foundation';
    UPDATE packages SET price_early_bird = 349000, price_normal = 500000 WHERE slug = 'pelatihan-intensif-osn-k-2026';
    UPDATE packages SET price_early_bird = 549000, price_normal = 899000 WHERE slug = 'paket-hemat-foundation-osn';
  END IF;
END $$;
