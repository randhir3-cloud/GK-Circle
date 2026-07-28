-- +migrate Up
-- EXAM-P5-T01: bind self-paced assessment_attempts to immutable test_snapshots
-- and enforce one IN_PROGRESS attempt per (quiz, user).

ALTER TABLE assessment_attempts
    ADD COLUMN IF NOT EXISTS test_snapshot_id uuid;

ALTER TABLE assessment_attempts
    DROP CONSTRAINT IF EXISTS assessment_attempts_test_snapshot_fk;

ALTER TABLE assessment_attempts
    ADD CONSTRAINT assessment_attempts_test_snapshot_fk
        FOREIGN KEY (test_snapshot_id) REFERENCES test_snapshots (id) ON DELETE RESTRICT;

-- Application create path always sets test_snapshot_id. Column stays nullable so
-- additive apply succeeds if any pre-T01 rows exist without a snapshot.
CREATE INDEX IF NOT EXISTS assessment_attempts_snapshot_id_idx
    ON assessment_attempts (test_snapshot_id);

CREATE INDEX IF NOT EXISTS assessment_attempts_quiz_user_status_idx
    ON assessment_attempts (quiz_id, user_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS assessment_attempts_one_active_uidx
    ON assessment_attempts (quiz_id, user_id)
    WHERE status = 'IN_PROGRESS';
