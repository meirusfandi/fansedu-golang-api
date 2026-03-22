-- Support collective purchase metadata (guru/instructor) and quantity pricing.

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS unit_price INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS subtotal INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS unique_code INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS is_collective BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS students_json JSONB;

UPDATE orders
SET
  quantity = 1,
  unit_price = COALESCE(NULLIF(unit_price, 0), COALESCE(normal_price, total_price)),
  subtotal = COALESCE(NULLIF(subtotal, 0), total_price),
  unique_code = COALESCE(NULLIF(unique_code, 0), 0),
  is_collective = COALESCE(is_collective, false)
WHERE quantity <= 0 OR unit_price = 0 OR subtotal = 0;

COMMENT ON COLUMN orders.quantity IS 'Jumlah item/siswa dalam satu order (siswa normal = 1)';
COMMENT ON COLUMN orders.unit_price IS 'Harga per item setelah promo (tanpa unique code)';
COMMENT ON COLUMN orders.subtotal IS 'unit_price * quantity (tanpa unique code)';
COMMENT ON COLUMN orders.unique_code IS 'Kode unik nominal transfer, ditambahkan 1x per order';
COMMENT ON COLUMN orders.is_collective IS 'True jika pembelian kolektif (umumnya role guru/instructor)';
COMMENT ON COLUMN orders.students_json IS 'Metadata siswa kolektif (array objek nama/email/user_id)';
