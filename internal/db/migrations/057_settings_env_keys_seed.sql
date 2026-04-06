-- Seed baris pengaturan opsional (override env). Nilai kosong = pakai env / default kode.
-- Admin bisa isi lewat CRUD /api/v1/admin/settings; restart API agar override diterapkan.

INSERT INTO settings (key, slug, value, description) VALUES
  ('JWT_SECRET', 'env-jwt-secret', NULL, 'Override JWT; jika kosong pakai env JWT_SECRET. Wajib kuat di production.'),
  ('OPENAI_API_KEY', 'env-openai-api-key', NULL, 'Override OpenAI untuk feedback AI.'),
  ('APP_URL', 'env-app-url', NULL, 'URL frontend (link email, checkout). Contoh: https://app.fansedu.com'),
  ('ADMIN_PASSWORD_BYPASS_KEY', 'env-admin-password-bypass-key', NULL, 'Opsional: emergency reset password admin.'),
  ('MIGRATE_BYPASS_KEY', 'env-migrate-bypass-key', NULL, 'Opsional: kunci emergency migrate via API.'),
  ('REDIS_URL', 'env-redis-url', NULL, 'Override Redis (cache geo, leaderboard, dll.).'),
  ('GEO_UPSTREAM_BASE_URL', 'env-geo-upstream-base-url', NULL, 'Upstream API wilayah Indonesia (emsifa).'),
  ('GEO_CACHE_TTL_SECONDS', 'env-geo-cache-ttl-seconds', NULL, 'TTL cache geo di Redis (angka, detik).'),
  ('LEADERBOARD_CACHE_TTL_SECONDS', 'env-leaderboard-cache-ttl-seconds', NULL, 'Legacy TTL leaderboard (detik).'),
  ('SCHOOL_LIST_CACHE_SECONDS', 'env-school-list-cache-seconds', NULL, 'TTL cache GET /schools (detik).'),
  ('PACKAGES_LIST_CACHE_SECONDS', 'env-packages-list-cache-seconds', NULL, 'TTL cache GET /packages (detik).'),
  ('SMTP_HOST', 'env-smtp-host', NULL, 'Host SMTP (mis. smtp-relay.brevo.com).'),
  ('SMTP_PORT', 'env-smtp-port', NULL, 'Port SMTP (mis. 587).'),
  ('SMTP_USER', 'env-smtp-user', NULL, 'User SMTP; kosong = pakai SMTP_FROM.'),
  ('SMTP_PASSWORD', 'env-smtp-password', NULL, 'Password SMTP / Brevo.'),
  ('BREVO_SMTP_KEY', 'env-brevo-smtp-key', NULL, 'Alias Brevo; jika diisi mengisi SMTP password.'),
  ('SMTP_FROM', 'env-smtp-from', NULL, 'Alamat From email.'),
  ('MIDTRANS_SERVER_KEY', 'env-midtrans-server-key', NULL, 'Server key Midtrans (Snap + verifikasi webhook).'),
  ('MIDTRANS_IS_PRODUCTION', 'env-midtrans-is-production', NULL, 'true = production Midtrans, false = sandbox.'),
  ('MIDTRANS_SNAP_BASE_URL', 'env-midtrans-snap-base-url', NULL, 'Opsional: override URL Snap API.')
ON CONFLICT (key) DO NOTHING;
