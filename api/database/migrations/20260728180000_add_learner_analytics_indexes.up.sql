-- +migrate Up

CREATE INDEX IF NOT EXISTS idx_attempts_user_submitted
  ON assessment_attempts(user_id, submitted_at DESC, id DESC)
  WHERE submitted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_attempts_user_quiz_submitted
  ON assessment_attempts(user_id, quiz_id, submitted_at DESC)
  WHERE submitted_at IS NOT NULL;
