-- +migrate Up
-- EXAM-P4-T04: immutable test composition snapshot contract (hooks for future attempts).

CREATE TABLE IF NOT EXISTS test_snapshots (
    id uuid PRIMARY KEY,
    quiz_id uuid NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
    created_by varchar(36) NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'CREATED',
    source_collection_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    question_count integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT test_snapshots_status_check
        CHECK (status IN ('CREATED'))
);

CREATE INDEX IF NOT EXISTS test_snapshots_quiz_id_created_at_idx
    ON test_snapshots (quiz_id, created_at DESC);

CREATE TABLE IF NOT EXISTS test_snapshot_items (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES test_snapshots (id) ON DELETE CASCADE,
    position integer NOT NULL,
    collection_id uuid,
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
    CONSTRAINT test_snapshot_items_snapshot_position_unique
        UNIQUE (snapshot_id, position),
    CONSTRAINT test_snapshot_items_snapshot_question_unique
        UNIQUE (snapshot_id, question_id)
);

CREATE INDEX IF NOT EXISTS test_snapshot_items_snapshot_id_idx
    ON test_snapshot_items (snapshot_id, position);
