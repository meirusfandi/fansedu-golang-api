-- Seed: create LMS courses from landing packages if missing.
-- Tujuan: memastikan setiap program (packages.slug) punya course backing
-- sehingga checkout (yang bergantung pada tabel courses) tidak 404.
--
-- Idempotent: hanya membuat course untuk slug yang belum ada di courses.

INSERT INTO courses (id, title, slug, description, price, thumbnail, subject_id, created_by, created_at, updated_at)
SELECT
  gen_random_uuid()                                    AS id,
  p.name                                               AS title,
  p.slug                                               AS slug,
  p.short_description                                  AS description,
  COALESCE(
    NULLIF(REGEXP_REPLACE(COALESCE(p.price_early_bird::text, p.price_normal::text, '0'), '[^0-9]', '', 'g'), '')::BIGINT,
    0
  )                                                    AS price,
  NULL                                                 AS thumbnail,
  NULL                                                 AS subject_id,
  NULL                                                 AS created_by,
  NOW()                                                AS created_at,
  NOW()                                                AS updated_at
FROM packages p
WHERE NOT EXISTS (
  SELECT 1 FROM courses c WHERE c.slug = p.slug
);

