-- +migrate Down

DROP TABLE IF EXISTS quiz_result_release_audit_logs;

ALTER TABLE quizzes
  DROP COLUMN IF EXISTS results_released_at;
