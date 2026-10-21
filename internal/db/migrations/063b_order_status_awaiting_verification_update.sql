-- Data migration for 063_order_status_awaiting_verification.sql.
-- Must run in a SEPARATE transaction from ALTER TYPE ADD VALUE (PG error 55P04 otherwise).

UPDATE orders
SET status = 'awaiting_verification'::order_status,
    updated_at = NOW()
WHERE status = 'pending'::order_status
  AND payment_proof_url IS NOT NULL;