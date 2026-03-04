-- Course content: module, quiz, test per kelas
DO $$ BEGIN
  CREATE TYPE course_content_type AS ENUM ('module', 'quiz', 'test');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS course_contents (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id   UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  title       VARCHAR(500) NOT NULL,
  description TEXT,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  type        course_content_type NOT NULL,
  content     JSONB,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_course_contents_course ON course_contents (course_id);
DROP TRIGGER IF EXISTS course_contents_updated_at ON course_contents;
CREATE TRIGGER course_contents_updated_at BEFORE UPDATE ON course_contents
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- Payments (pembayaran user: kelas, subscription, dll)
DO $$ BEGIN
  CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
DO $$ BEGIN
  CREATE TYPE payment_type AS ENUM ('course_purchase', 'subscription', 'tryout', 'other');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS payments (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  amount_cents  INTEGER NOT NULL,
  currency      VARCHAR(3) NOT NULL DEFAULT 'IDR',
  status        payment_status NOT NULL DEFAULT 'pending',
  type          payment_type NOT NULL DEFAULT 'course_purchase',
  reference_id  UUID,
  description   TEXT,
  paid_at       TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user ON payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status);
CREATE INDEX IF NOT EXISTS idx_payments_created ON payments (created_at);
DROP TRIGGER IF EXISTS payments_updated_at ON payments;
CREATE TRIGGER payments_updated_at BEFORE UPDATE ON payments
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
