-- +migrate Down

DROP INDEX IF EXISTS idx_analytics_quiz_timeline;
DROP INDEX IF EXISTS idx_attempt_answers_attempt_question;
DROP INDEX IF EXISTS idx_attempts_quiz_user_status;
DROP INDEX IF EXISTS idx_attempts_quiz_submitted;
