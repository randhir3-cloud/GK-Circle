-- +migrate Down
-- EXAM-P5-T04: remove attempt-linked snapshot items and frozen scoring columns.

DROP TABLE IF EXISTS assessment_attempt_snapshot_items;

ALTER TABLE assessment_attempts
    DROP COLUMN IF EXISTS expected_max_score;

ALTER TABLE assessment_attempts
    DROP COLUMN IF EXISTS negative_marks_per_question;
