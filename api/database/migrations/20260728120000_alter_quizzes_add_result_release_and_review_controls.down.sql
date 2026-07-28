-- +migrate Down

ALTER TABLE quizzes
  DROP CONSTRAINT IF EXISTS quizzes_result_release_policy_check,
  DROP COLUMN IF EXISTS result_release_policy,
  DROP COLUMN IF EXISTS results_released,
  DROP COLUMN IF EXISTS results_scheduled_at,
  DROP COLUMN IF EXISTS show_score,
  DROP COLUMN IF EXISTS show_pass_fail,
  DROP COLUMN IF EXISTS show_correctness,
  DROP COLUMN IF EXISTS show_explanations;
