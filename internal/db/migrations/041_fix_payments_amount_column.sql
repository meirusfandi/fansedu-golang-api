-- Fix legacy schema mismatch where payments.amount column may be missing.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'payments'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'amount'
    ) THEN
      ALTER TABLE public.payments ADD COLUMN amount INTEGER NOT NULL DEFAULT 0;
    END IF;

    -- Try to backfill from common legacy column names when present.
    IF EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'total_amount'
    ) THEN
      EXECUTE 'UPDATE public.payments SET amount = COALESCE(NULLIF(amount, 0), COALESCE(total_amount, 0))';
    ELSIF EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'gross_amount'
    ) THEN
      EXECUTE 'UPDATE public.payments SET amount = COALESCE(NULLIF(amount, 0), COALESCE(gross_amount, 0))';
    ELSIF EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = 'payments' AND column_name = 'total'
    ) THEN
      EXECUTE 'UPDATE public.payments SET amount = COALESCE(NULLIF(amount, 0), COALESCE(total, 0))';
    END IF;
  END IF;
END
$$;
