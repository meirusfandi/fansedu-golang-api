-- Normalize payments amount column naming: remove *_cents suffix.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'payments'
  ) THEN
    -- Case A: only amount_cents exists -> rename directly.
    IF EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'amount_cents'
    ) AND NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'amount'
    ) THEN
      ALTER TABLE public.payments RENAME COLUMN amount_cents TO amount;
    END IF;

    -- Case B: both exist -> backfill amount from amount_cents then drop amount_cents.
    IF EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'amount_cents'
    ) AND EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'amount'
    ) THEN
      UPDATE public.payments
      SET amount = COALESCE(NULLIF(amount, 0), COALESCE(amount_cents, 0));

      ALTER TABLE public.payments DROP COLUMN amount_cents;
    END IF;
  END IF;
END
$$;
