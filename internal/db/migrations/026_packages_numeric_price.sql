-- Packages: simpan nominal harga dalam bentuk angka (rupiah).

ALTER TABLE packages ADD COLUMN IF NOT EXISTS price_early_bird BIGINT;
ALTER TABLE packages ADD COLUMN IF NOT EXISTS price_normal BIGINT;

-- Nominal untuk seed packages (slugs dikenal) - nilai dalam rupiah
UPDATE packages SET price_early_bird = 249000, price_normal = 399000 WHERE slug = 'algorithm-programming-foundation';
UPDATE packages SET price_early_bird = 349000, price_normal = 500000 WHERE slug = 'pelatihan-intensif-osn-k-2026';
UPDATE packages SET price_early_bird = 549000, price_normal = 899000 WHERE slug = 'paket-hemat-foundation-osn';

COMMENT ON COLUMN packages.price_early_bird IS 'Harga early bird (rupiah)';
COMMENT ON COLUMN packages.price_normal IS 'Harga normal (rupiah)';
