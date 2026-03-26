-- Pastikan orders.package_id ada (fix ERROR 42703 jika 043 belum pernah jalan / deploy lama).
-- Idempotent: aman dipanggil berulang.

ALTER TABLE orders ADD COLUMN IF NOT EXISTS package_id UUID REFERENCES packages(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_orders_package_id ON orders(package_id) WHERE package_id IS NOT NULL;
