-- Add intermediate order status after payment proof upload.
-- Flow: pending -> awaiting_verification -> paid|failed

ALTER TYPE order_status ADD VALUE IF NOT EXISTS 'awaiting_verification';

UPDATE orders
SET status = 'awaiting_verification'::order_status,
    updated_at = NOW()
WHERE status = 'pending'::order_status
  AND payment_proof_url IS NOT NULL;
