-- LMS Checkout: frictionless purchase (orders, order_items, course slug/price, payment gateway)
-- Supports: GET courses by slug, checkout without login, Midtrans/Stripe webhook

-- 1. Users: password nullable for auto-created accounts at checkout
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- 2. Courses: slug, price, thumbnail (instructor = created_by)
ALTER TABLE courses ADD COLUMN IF NOT EXISTS slug VARCHAR(255);
ALTER TABLE courses ADD COLUMN IF NOT EXISTS price INTEGER NOT NULL DEFAULT 0;
ALTER TABLE courses ADD COLUMN IF NOT EXISTS thumbnail VARCHAR(512);
CREATE UNIQUE INDEX IF NOT EXISTS idx_courses_slug ON courses (slug) WHERE slug IS NOT NULL;

-- 3. Order status enum
DO $$ BEGIN
  CREATE TYPE order_status AS ENUM ('pending', 'paid', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 4. Orders
CREATE TABLE IF NOT EXISTS orders (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id            UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  status             order_status NOT NULL DEFAULT 'pending',
  total_price        INTEGER NOT NULL DEFAULT 0,
  payment_method     VARCHAR(50),
  payment_reference  VARCHAR(255),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_orders_created ON orders (created_at);
DROP TRIGGER IF EXISTS orders_updated_at ON orders;
CREATE TRIGGER orders_updated_at BEFORE UPDATE ON orders
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- 5. Order items
CREATE TABLE IF NOT EXISTS order_items (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id   UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
  course_id  UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  price      INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_course ON order_items (course_id);

-- 6. Payments: link to order + gateway fields (existing payments table extended)
ALTER TABLE payments ADD COLUMN IF NOT EXISTS order_id UUID REFERENCES orders (id) ON DELETE SET NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS gateway VARCHAR(50);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS transaction_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments (order_id);

-- 7. instructor role (alias for guru / BuildWithAngga-style)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    WHERE t.typname = 'user_role' AND e.enumlabel = 'instructor') THEN
    ALTER TYPE user_role ADD VALUE 'instructor';
  END IF;
END $$;

COMMENT ON TABLE orders IS 'Checkout orders; pending → paid after gateway webhook';
COMMENT ON TABLE order_items IS 'Line items per order (one course per item)';
COMMENT ON COLUMN payments.order_id IS 'Set when payment is for an order (checkout flow)';
