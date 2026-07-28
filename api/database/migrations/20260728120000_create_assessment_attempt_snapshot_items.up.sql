-- +migrate Up
-- EXAM-P5-T04: attempt-linked resolved snapshot persistence + frozen scoring config.

ALTER TABLE assessment_attempts
    ADD COLUMN IF NOT EXISTS negative_marks_per_question numeric(4,2) NOT NULL DEFAULT 0;

ALTER TABLE assessment_attempts
    ADD COLUMN IF NOT EXISTS expected_max_score numeric(8,2);

CREATE TABLE IF NOT EXISTS assessment_attempt_snapshot_items (
    id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL,
    snapshot_item_id uuid NOT NULL,
    position integer NOT NULL,
    question_id uuid NOT NULL,
    lineage_id uuid NOT NULL,
    revision_number integer NOT NULL,
    question text NOT NULL,
    type integer NOT NULL,
    options jsonb NOT NULL,
    answers jsonb NOT NULL,
    official_answer jsonb NOT NULL,
    authoritative_answer jsonb NOT NULL,
    answer_review_status varchar(20) NOT NULL,
    points smallint,
    duration_in_seconds integer,
    question_media varchar(10),
    options_media varchar(10),
    resource text,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT assessment_attempt_snapshot_items_attempt_fk
        FOREIGN KEY (attempt_id) REFERENCES assessment_attempts (id) ON DELETE RESTRICT,
    CONSTRAINT assessment_attempt_snapshot_items_snapshot_item_fk
        FOREIGN KEY (snapshot_item_id) REFERENCES test_snapshot_items (id) ON DELETE RESTRICT,
    CONSTRAINT assessment_attempt_snapshot_items_attempt_position_uidx
        UNIQUE (attempt_id, position),
    CONSTRAINT assessment_attempt_snapshot_items_attempt_question_uidx
        UNIQUE (attempt_id, question_id)
);

CREATE INDEX IF NOT EXISTS assessment_attempt_snapshot_items_attempt_id_idx
    ON assessment_attempt_snapshot_items (attempt_id, position);
