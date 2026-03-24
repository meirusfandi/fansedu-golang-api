-- Package ↔ kelas (LMS): satu paket landing bisa berisi banyak course.
-- Setelah pembayaran, siswa di-enroll ke semua course di package_courses.

CREATE TABLE IF NOT EXISTS package_courses (
  package_id UUID NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
  course_id  UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  sort_order INT NOT NULL DEFAULT 0,
  PRIMARY KEY (package_id, course_id)
);
CREATE INDEX IF NOT EXISTS idx_package_courses_course ON package_courses(course_id);

COMMENT ON TABLE package_courses IS 'Kelas LMS yang termasuk dalam paket landing; urutan via sort_order';

ALTER TABLE orders ADD COLUMN IF NOT EXISTS package_id UUID REFERENCES packages(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_orders_package_id ON orders(package_id) WHERE package_id IS NOT NULL;

-- Backfill: satu course per paket jika slug packages = slug courses (kompatibel dengan flow lama).
INSERT INTO package_courses (package_id, course_id, sort_order)
SELECT p.id, c.id, 0
FROM packages p
INNER JOIN courses c ON c.slug = p.slug
ON CONFLICT (package_id, course_id) DO NOTHING;
