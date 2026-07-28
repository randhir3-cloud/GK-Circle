-- +migrate Up

ALTER TABLE quiz_result_release_audit_logs
  ADD COLUMN IF NOT EXISTS correlation_id varchar(128);

CREATE INDEX IF NOT EXISTS idx_quiz_result_release_audit_logs_quiz_created
  ON quiz_result_release_audit_logs (quiz_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_quiz_result_release_audit_logs_actor_created
  ON quiz_result_release_audit_logs (actor_id, created_at DESC);
