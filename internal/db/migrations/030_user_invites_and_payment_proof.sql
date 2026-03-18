-- User invites: token untuk guest checkout agar bisa daftar akun (set password)
CREATE TABLE IF NOT EXISTS user_invites (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  order_id   UUID REFERENCES orders (id) ON DELETE SET NULL,
  email      VARCHAR(255) NOT NULL,
  name       VARCHAR(255),
  token      VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_invites_token ON user_invites (token);
CREATE INDEX IF NOT EXISTS idx_user_invites_user ON user_invites (user_id);
CREATE INDEX IF NOT EXISTS idx_user_invites_order ON user_invites (order_id);

-- Bukti pembayaran per order (upload proof)
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_proof_url VARCHAR(1024);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_proof_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS sender_account_no VARCHAR(100);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS sender_name VARCHAR(255);

COMMENT ON TABLE user_invites IS 'Invite token untuk guest checkout: link registrasi di email';
COMMENT ON COLUMN orders.payment_proof_url IS 'URL/path bukti transfer yang di-upload user';
