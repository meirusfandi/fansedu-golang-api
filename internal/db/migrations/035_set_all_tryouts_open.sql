-- Make all tryout sessions available as "open"
-- One-time / idempotent operation: set status='open' for anything else.
UPDATE tryout_sessions
SET status = 'open'::tryout_status
WHERE status IS DISTINCT FROM 'open'::tryout_status;

