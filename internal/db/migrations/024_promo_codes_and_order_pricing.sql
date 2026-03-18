-- Promo codes and order pricing (normal, discount, final, confirmation code)

-- 1. Promo codes table
CREATE TABLE IF NOT EXISTS promo_codes (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code          VARCHAR(64) NOT NULL UNIQUE,
  discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
  discount_value INTEGER NOT NULL,
  valid_from    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  valid_until    TIMESTAMPTZ,
  max_uses      INTEGER,
  used_count    INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_promo_codes_code ON promo_codes (LOWER(code));
CREATE INDEX IF NOT EXISTS idx_promo_codes_valid ON promo_codes (valid_from, valid_until);

-- 2. Orders: add pricing breakdown and 3-digit confirmation code
ALTER TABLE orders ADD COLUMN IF NOT EXISTS normal_price INTEGER NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS promo_code VARCHAR(64);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_percent NUMERIC(5,2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS confirmation_code VARCHAR(3);

COMMENT ON COLUMN orders.normal_price IS 'Harga normal sebelum promo (rupiah)';
COMMENT ON COLUMN orders.promo_code IS 'Kode promo yang dipakai (jika ada)';
COMMENT ON COLUMN orders.discount IS 'Potongan harga (rupiah)';
COMMENT ON COLUMN orders.discount_percent IS 'Persen diskon (0-100)';
COMMENT ON COLUMN orders.confirmation_code IS '3 digit unik untuk validasi konfirmasi pembayaran';
COMMENT ON COLUMN orders.total_price IS 'Harga final setelah diskon (rupiah)';
