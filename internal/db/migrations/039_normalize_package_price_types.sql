-- Normalize packages price columns to BIGINT (rupiah).
-- Handles legacy schemas where price_early_bird / price_normal are VARCHAR.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'packages'
  ) THEN
    ALTER TABLE packages
      ALTER COLUMN price_early_bird TYPE BIGINT
      USING NULLIF(REGEXP_REPLACE(COALESCE(price_early_bird::text, ''), '[^0-9]', '', 'g'), '')::BIGINT;

    ALTER TABLE packages
      ALTER COLUMN price_normal TYPE BIGINT
      USING NULLIF(REGEXP_REPLACE(COALESCE(price_normal::text, ''), '[^0-9]', '', 'g'), '')::BIGINT;
  END IF;
END $$;
