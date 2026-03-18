-- Seed: contoh kode promo untuk testing (opsional)
-- discount_type: 'percent' = diskon persen, 'fixed' = potongan tetap (rupiah)
-- discount_value: untuk percent 1-100, untuk fixed = jumlah rupiah (e.g. 50000 = Rp 50.000)

INSERT INTO promo_codes (id, code, discount_type, discount_value, valid_from, valid_until, max_uses, used_count, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'GRATIS100', 'percent', 100, NOW(), NOW() + INTERVAL '1 year', 1000, 0, NOW(), NOW()),
  (gen_random_uuid(), 'DISKON50', 'percent', 50, NOW(), NOW() + INTERVAL '1 year', NULL, 0, NOW(), NOW()),
  (gen_random_uuid(), 'POTONG50K', 'fixed', 50000, NOW(), NOW() + INTERVAL '1 year', NULL, 0, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;
