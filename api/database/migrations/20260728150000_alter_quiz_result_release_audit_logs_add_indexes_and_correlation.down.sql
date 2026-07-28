-- +migrate Down

DROP INDEX IF EXISTS idx_quiz_result_release_audit_logs_actor_created;
DROP INDEX IF EXISTS idx_quiz_result_release_audit_logs_quiz_created;

ALTER TABLE quiz_result_release_audit_logs
  DROP COLUMN IF EXISTS correlation_id;
