-- =============================================================================
-- Schema Database: Konten Landing Page (Admin-Managed)
-- =============================================================================
-- Untuk PostgreSQL. Untuk MySQL: ganti TIMESTAMP WITH TIME ZONE → DATETIME,
-- DEFAULT NOW() → CURRENT_TIMESTAMP, dan BOOLEAN → TINYINT(1).
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. Site Settings (key-value: hero, urgency, CTA, WA, nav, dll.)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS site_settings (
  key   VARCHAR(100) PRIMARY KEY,
  value TEXT NOT NULL
);

-- Contoh key yang dipakai frontend (admin isi via panel):
-- Hero
--   hero_badge              → badge kecil di atas headline (e.g. "Program Persiapan OSN Informatika 2026")
--   hero_headline_1         → "Persiapan OSN Informatika 2026"
--   hero_headline_2         → "Mulai dari Dasar sampai Lolos OSN-K"
--   hero_subheadline        → paragraf di bawah headline
--   hero_cta_primary        → "Daftar Sekarang"
--   hero_cta_secondary      → "Tanya Program"
--   hero_cta_tertiary       → "Pelajari Lebih Lanjut"
-- Urgency (section CTA & kartu program)
--   urgency_batch           → "Batch April 2026"
--   urgency_quota_max       → "30"
--   urgency_early_bird_end  → "2026-03-17T23:59:59+07:00" (ISO 8601)
-- CTA section (Siap Lolos OSN-K?)
--   cta_headline            → "Siap Lolos"
--   cta_headline_accent     → "OSN-K 2026?"
--   cta_subtitle            → paragraf
--   cta_button_primary      → "Daftar Sekarang"
--   cta_button_secondary    → "Tanya Program"
-- WA & links
--   wa_number               → "6285121277161"
--   wa_template_navbar      → pesan WA untuk tombol navbar
--   wa_template_hero        → pesan WA untuk CTA hero
--   wa_template_tanya       → pesan WA untuk "Tanya Program"
--   wa_template_request     → pesan WA untuk Request Bidang
--   wa_template_contact     → pesan WA untuk Kontak
--   wa_template_float       → pesan WA untuk tombol float
--   register_url            → "https://app.fansedu.web.id/register"
-- Solusi section
--   solusi_badge            → "Solusi Kami"
--   solusi_headline         → "Dari Masalah ke Lolos OSN-K"
--   solusi_paragraph_1      → paragraf pertama
--   solusi_paragraph_2      → paragraf (tim pengajar)
--   solusi_logo_url         → "/fansedu.png"
-- Masalah section
--   masalah_badge           → "Yang Sering Dihadapi"
--   masalah_headline        → "Masalah Siswa Persiapan OSN"
--   masalah_subtitle        → paragraf intro
-- Fitur section
--   features_badge          → "Fitur Program"
--   features_headline      → "Fitur Program FansEdu"
-- Tryout section
--   tryout_badge            → "Free Tryout"
--   tryout_headline         → "Coba TryOut OSN Gratis"
--   tryout_subtitle         → paragraf
--   tryout_cta_text         → "Ikuti TryOut Gratis"
--   tryout_cta_href         → "#/tryout-info"
-- Request bidang
--   request_badge           → "Request Bidang Lainnya"
--   request_headline        → "Ingin Request Bidang Lainnya?"
--   request_subtitle        → paragraf
--   request_button          → "Kirim Request"
-- Contact section
--   contact_badge           → "Hubungi Kami"
--   contact_headline        → "Siap Memulai Perjalanan?"
--   contact_subtitle        → paragraf
-- Footer
--   footer_logo_text        → "Fansedu Informatic Olympiad"
--   footer_tagline         → paragraf
--   footer_copyright        → "2026 Fansedu ..."
-- Visual proof (section Lihat Sendiri)
--   visual_headline         → "Lihat Sendiri Pengalaman Belajar"
--   visual_leaderboard_url  → "/leaderboard-to.png"
--   visual_leaderboard_title→ "Leaderboard Nasional"
--   visual_leaderboard_desc → "Peringkat tryout antar peserta"
--   visual_dashboard_url    → "/dashboard-siswa.png"
--   visual_dashboard_title  → "Dashboard Peserta"
--   visual_dashboard_desc   → "Progress dan akses materi"
--   visual_kelas_url        → "/kelas-osn.png"
--   visual_kelas_title      → "Kelas"
--   visual_kelas_desc       → "Sesi belajar dan materi OSN"


-- -----------------------------------------------------------------------------
-- 2. Hero badges (3 poin di bawah subheadline: ✔ Mentor berpengalaman, dll.)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS hero_badges (
  id         VARCHAR(36) PRIMARY KEY,
  label      VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_hero_badges_sort ON hero_badges(sort_order);


-- -----------------------------------------------------------------------------
-- 3. Social proof (section angka: 20+, 2x, 100%, sejak 2014)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS social_proof_stats (
  id         VARCHAR(36) PRIMARY KEY,
  value      VARCHAR(100) NOT NULL,
  label      VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_social_proof_stats_sort ON social_proof_stats(sort_order);


-- -----------------------------------------------------------------------------
-- 4. Masalah siswa (section pain points: judul + deskripsi)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS masalah_items (
  id          VARCHAR(36) PRIMARY KEY,
  title       VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  sort_order  INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_masalah_items_sort ON masalah_items(sort_order);


-- -----------------------------------------------------------------------------
-- 5. Solusi highlights (bullet list: Kurikulum terstruktur, Tryout nasional, dll.)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS solusi_highlights (
  id         VARCHAR(36) PRIMARY KEY,
  label      VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_solusi_highlights_sort ON solusi_highlights(sort_order);


-- -----------------------------------------------------------------------------
-- 6. Features (section Fitur Program: judul + deskripsi)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS features (
  id          VARCHAR(36) PRIMARY KEY,
  title       VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  sort_order  INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_features_sort ON features(sort_order);


-- -----------------------------------------------------------------------------
-- 7. Packages / Program (paket yang sedang dibuka)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS packages (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                 VARCHAR(255) NOT NULL,
  slug                 VARCHAR(255) NOT NULL UNIQUE,
  short_description    TEXT,
  price_display        VARCHAR(100),
  price_early_bird     VARCHAR(100),
  price_normal         VARCHAR(100),
  cta_label            VARCHAR(255) NOT NULL DEFAULT 'Daftar',
  wa_message_template  TEXT,
  cta_url              VARCHAR(512),
  is_open              BOOLEAN NOT NULL DEFAULT true,
  is_bundle            BOOLEAN NOT NULL DEFAULT false,
  bundle_subtitle      VARCHAR(255),
  durasi               VARCHAR(100),
  materi               JSONB DEFAULT '[]',
  fasilitas            JSONB DEFAULT '[]',
  bonus                JSONB DEFAULT '[]',
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE packages IS 'Landing page packages (Program yang Sedang Dibuka)';
CREATE INDEX IF NOT EXISTS idx_packages_slug ON packages(slug);
CREATE INDEX IF NOT EXISTS idx_packages_is_open ON packages(is_open);


-- -----------------------------------------------------------------------------
-- 8. Testimonials (section Testimoni: quote, author, role)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS testimonials (
  id         VARCHAR(36) PRIMARY KEY,
  quote      TEXT NOT NULL,
  author     VARCHAR(255) NOT NULL,
  role       VARCHAR(255) NOT NULL,
  is_active  BOOLEAN NOT NULL DEFAULT true,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_testimonials_active_sort ON testimonials(is_active, sort_order);


-- -----------------------------------------------------------------------------
-- 9. Tryout steps (4 langkah: Tryout gratis, Lihat ranking, dll.)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tryout_steps (
  id          VARCHAR(36) PRIMARY KEY,
  step_number INT NOT NULL,
  title       VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  sort_order  INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_tryout_steps_sort ON tryout_steps(sort_order);


-- -----------------------------------------------------------------------------
-- 10. Contact links (WhatsApp, Instagram, TikTok, YouTube)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS contact_links (
  id         VARCHAR(36) PRIMARY KEY,
  title      VARCHAR(100) NOT NULL,
  subtitle   VARCHAR(255),
  href       VARCHAR(1024) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_contact_links_sort ON contact_links(sort_order);


-- -----------------------------------------------------------------------------
-- 11. Nav menu (opsional: urutan & label link navbar/footer)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS nav_items (
  id         VARCHAR(36) PRIMARY KEY,
  label      VARCHAR(100) NOT NULL,
  anchor     VARCHAR(100) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  is_visible BOOLEAN NOT NULL DEFAULT true
);
CREATE INDEX IF NOT EXISTS idx_nav_items_visible_sort ON nav_items(is_visible, sort_order);


-- -----------------------------------------------------------------------------
-- 12. About highlights (jika masih dipakai; section Solusi punya solusi_highlights)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS about_highlights (
  id         VARCHAR(36) PRIMARY KEY,
  label      VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_about_highlights_sort ON about_highlights(sort_order);


-- -----------------------------------------------------------------------------
-- 13. Services (section Layanan — jika dipakai)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS services (
  id          VARCHAR(36) PRIMARY KEY,
  title       VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  sort_order  INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_services_sort ON services(sort_order);


-- =============================================================================
-- SEED / CONTOH (opsional; bisa dijalankan terpisah)
-- =============================================================================
/*
-- Hero badges
INSERT INTO hero_badges (id, label, sort_order) VALUES
  (gen_random_uuid()::text, 'Mentor berpengalaman', 1),
  (gen_random_uuid()::text, '2x Tryout Nasional', 2),
  (gen_random_uuid()::text, 'Dashboard Ranking', 3);

-- Social proof
INSERT INTO social_proof_stats (id, value, label, sort_order) VALUES
  (gen_random_uuid()::text, '20+', 'paket pembahasan soal', 1),
  (gen_random_uuid()::text, '2x', 'Tryout nasional per batch', 2),
  (gen_random_uuid()::text, '100%', 'Pembahasan mendalam', 3),
  (gen_random_uuid()::text, 'sejak 2014', 'Pengalaman bimbingan OSN', 4);

-- Masalah
INSERT INTO masalah_items (id, title, description, sort_order) VALUES
  (gen_random_uuid()::text, 'Bingung mulai dari mana', 'Materi OSN luas, tidak tahu prioritas dan urutan belajar yang efektif.', 1),
  (gen_random_uuid()::text, 'Materi tidak terstruktur', 'Belajar dari banyak sumber tanpa kurikulum jelas dan pembahasan mendalam.', 2);

-- Solusi highlights
INSERT INTO solusi_highlights (id, label, sort_order) VALUES
  (gen_random_uuid()::text, 'Kurikulum terstruktur', 1),
  (gen_random_uuid()::text, 'Tryout nasional', 2),
  (gen_random_uuid()::text, 'Pembahasan mendalam', 3),
  (gen_random_uuid()::text, 'Mentor berpengalaman', 4),
  (gen_random_uuid()::text, 'Dashboard ranking', 5),
  (gen_random_uuid()::text, 'Rekaman & arsip lengkap', 6);

-- Tryout steps
INSERT INTO tryout_steps (id, step_number, title, description, sort_order) VALUES
  (gen_random_uuid()::text, 1, 'Tryout gratis', 'Ikuti TryOut OSN format resmi. Gratis, tanpa biaya.', 1),
  (gen_random_uuid()::text, 2, 'Lihat ranking', 'Cek peringkatmu di leaderboard nasional.', 2),
  (gen_random_uuid()::text, 3, 'Analisis hasil', 'Pahami kekuatan & area yang perlu ditingkatkan.', 3),
  (gen_random_uuid()::text, 4, 'Offer kelas', 'Dapat rekomendasi program yang sesuai dengan kebutuhanmu.', 4);

-- Nav
INSERT INTO nav_items (id, label, anchor, sort_order) VALUES
  (gen_random_uuid()::text, 'Tentang Kami', 'solusi', 1),
  (gen_random_uuid()::text, 'Tantangan', 'masalah', 2),
  (gen_random_uuid()::text, 'Fitur', 'features', 3),
  (gen_random_uuid()::text, 'Program', 'packages', 4),
  (gen_random_uuid()::text, 'Testimoni', 'testimoni', 5),
  (gen_random_uuid()::text, 'TryOut Gratis', 'tryout', 6),
  (gen_random_uuid()::text, 'Request Bidang', 'request', 7),
  (gen_random_uuid()::text, 'Kontak', 'contact', 8);
*/
