-- +migrate Down

DROP INDEX IF EXISTS quizzes_access_code_uidx;

ALTER TABLE quizzes
  DROP CONSTRAINT IF EXISTS quizzes_assessment_mode_check,
  DROP CONSTRAINT IF EXISTS quizzes_status_check,
  DROP COLUMN IF EXISTS assessment_mode,
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS published_at,
  DROP COLUMN IF EXISTS duration_seconds,
  DROP COLUMN IF EXISTS max_attempts,
  DROP COLUMN IF EXISTS access_code,
  DROP COLUMN IF EXISTS negative_marks_per_question,
  DROP COLUMN IF EXISTS show_leaderboard,
  DROP COLUMN IF EXISTS allow_answer_review;
