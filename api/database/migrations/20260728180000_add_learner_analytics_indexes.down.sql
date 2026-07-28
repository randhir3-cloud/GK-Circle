-- +migrate Down

DROP INDEX IF EXISTS idx_attempts_user_quiz_submitted;
DROP INDEX IF EXISTS idx_attempts_user_submitted;
