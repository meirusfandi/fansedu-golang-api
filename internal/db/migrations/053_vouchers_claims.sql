-- Voucher / promo: aktif, klaim per user, scope checkout kelas vs paket.

ALTER TABLE promo_codes
  ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE promo_codes
  ADD COLUMN IF NOT EXISTS requires_claim BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE promo_codes
  ADD COLUMN IF NOT EXISTS applies_to_courses BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE promo_codes
  ADD COLUMN IF NOT EXISTS applies_to_packages BOOLEAN NOT NULL DEFAULT TRUE;

COMMENT ON COLUMN promo_codes.is_active IS 'Admin bisa menonaktifkan tanpa hapus baris';
COMMENT ON COLUMN promo_codes.requires_claim IS 'TRUE: user wajib POST /vouchers/claim sebelum pakai di checkout';
COMMENT ON COLUMN promo_codes.applies_to_courses IS 'Bisa dipakai checkout satu kelas (Initiate)';
COMMENT ON COLUMN promo_codes.applies_to_packages IS 'Bisa dipakai checkout paket landing (InitiatePackage)';

-- Klaim voucher ke akun user (satu baris per user per promo).
CREATE TABLE IF NOT EXISTS voucher_claims (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  promo_code_id   UUID NOT NULL REFERENCES promo_codes (id) ON DELETE CASCADE,
  claimed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  used_at         TIMESTAMPTZ,
  order_id        UUID REFERENCES orders (id) ON DELETE SET NULL,
  UNIQUE (user_id, promo_code_id)
);

CREATE INDEX IF NOT EXISTS idx_voucher_claims_user_unused
  ON voucher_claims (user_id) WHERE used_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_voucher_claims_promo ON voucher_claims (promo_code_id);
