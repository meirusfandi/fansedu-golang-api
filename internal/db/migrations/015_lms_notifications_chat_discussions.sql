-- LMS: notifications, course chat, discussions, payment proof

-- Notifications (user notifications)
CREATE TABLE IF NOT EXISTS notifications (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  title      VARCHAR(500) NOT NULL,
  body       TEXT,
  type       VARCHAR(50) DEFAULT 'system',
  read_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read_at ON notifications (user_id, read_at);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at);

-- Course chat messages
CREATE TABLE IF NOT EXISTS course_messages (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id  UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  message    TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_course_messages_course ON course_messages (course_id);
CREATE INDEX IF NOT EXISTS idx_course_messages_created ON course_messages (course_id, created_at);

-- Course discussions (forum)
CREATE TABLE IF NOT EXISTS course_discussions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id  UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  title      VARCHAR(500) NOT NULL,
  body       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_course_discussions_course ON course_discussions (course_id);
CREATE INDEX IF NOT EXISTS idx_course_discussions_created ON course_discussions (course_id, created_at);

DROP TRIGGER IF EXISTS course_discussions_updated_at ON course_discussions;
CREATE TRIGGER course_discussions_updated_at BEFORE UPDATE ON course_discussions
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE IF NOT EXISTS course_discussion_replies (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  discussion_id UUID NOT NULL REFERENCES course_discussions (id) ON DELETE CASCADE,
  user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  body          TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_course_discussion_replies_discussion ON course_discussion_replies (discussion_id);
CREATE INDEX IF NOT EXISTS idx_course_discussion_replies_created ON course_discussion_replies (discussion_id, created_at);

DROP TRIGGER IF EXISTS course_discussion_replies_updated_at ON course_discussion_replies;
CREATE TRIGGER course_discussion_replies_updated_at BEFORE UPDATE ON course_discussion_replies
  FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- Payment proof (upload bukti transfer)
ALTER TABLE payments ADD COLUMN IF NOT EXISTS proof_url VARCHAR(1024);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS confirmed_by UUID REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS confirmed_at TIMESTAMPTZ;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS rejection_note TEXT;
