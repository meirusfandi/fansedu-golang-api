-- Add intermediate order status after payment proof upload.
-- Flow: pending -> awaiting_verification -> paid|failed
-- NOTE: ALTER TYPE ADD VALUE + usage in same implicit transaction causes PG error 55P04.
-- The UPDATE is split into a separate file (063b_order_status_awaiting_verification_update.sql)
-- so it runs in a different transaction.

ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'awaiting_verification';
