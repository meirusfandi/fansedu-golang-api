-- Update courses that have price = 0 or NULL from their corresponding packages
-- Uses early_bird price first, then normal price

UPDATE courses c
SET price = COALESCE(
  (SELECT COALESCE(p.price_early_bird, p.price_normal)
   FROM packages p
   WHERE p.slug = c.slug
   LIMIT 1),
  c.price
)
WHERE (c.price = 0 OR c.price IS NULL)
  AND c.slug IS NOT NULL
  AND EXISTS (SELECT 1 FROM packages p WHERE p.slug = c.slug);

-- Also update courses where slug matches package slug and price differs significantly
-- (in case previous lazy creation used wrong price)
UPDATE courses c
SET price = (SELECT COALESCE(p.price_early_bird, p.price_normal)
             FROM packages p
             WHERE p.slug = c.slug
             LIMIT 1)
WHERE c.slug IS NOT NULL
  AND c.price = 0
  AND EXISTS (
    SELECT 1 FROM packages p
    WHERE p.slug = c.slug
    AND COALESCE(p.price_early_bird, p.price_normal, 0) > 0
  );
