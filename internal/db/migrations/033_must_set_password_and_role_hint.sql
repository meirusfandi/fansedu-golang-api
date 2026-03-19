-- Add must_set_password flag to users for guest checkout flow
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_set_password BOOLEAN NOT NULL DEFAULT FALSE;

-- Add role_hint to orders for auto-create user with correct role
ALTER TABLE orders ADD COLUMN IF NOT EXISTS role_hint VARCHAR(20);

-- Add email column to orders for guest checkout (in case user_id is placeholder)
ALTER TABLE orders ADD COLUMN IF NOT EXISTS buyer_email VARCHAR(255);

-- Index for fast lookup
CREATE INDEX IF NOT EXISTS idx_orders_buyer_email ON orders(buyer_email);
CREATE INDEX IF NOT EXISTS idx_users_must_set_password ON users(must_set_password) WHERE must_set_password = TRUE;
