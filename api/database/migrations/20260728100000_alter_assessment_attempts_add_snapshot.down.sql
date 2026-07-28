-- +migrate Down
-- EXAM-P5-T01: remove snapshot binding and one-active index from assessment_attempts.

DROP INDEX IF EXISTS assessment_attempts_one_active_uidx;
DROP INDEX IF EXISTS assessment_attempts_quiz_user_status_idx;
DROP INDEX IF EXISTS assessment_attempts_snapshot_id_idx;

ALTER TABLE assessment_attempts
    DROP CONSTRAINT IF EXISTS assessment_attempts_test_snapshot_fk;

ALTER TABLE assessment_attempts
    DROP COLUMN IF EXISTS test_snapshot_id;
