-- +migrate Up
-- EXAM-P4-T01: Question Collection entities (STATIC membership + DYNAMIC filters).

CREATE TABLE IF NOT EXISTS question_collections (
    id uuid PRIMARY KEY,
    quiz_id uuid NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
    title text NOT NULL,
    kind varchar(16) NOT NULL,
    position integer NOT NULL DEFAULT 0,
    filter_json jsonb,
    created_by varchar(36) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT question_collections_kind_check
        CHECK (kind IN ('STATIC', 'DYNAMIC')),
    CONSTRAINT question_collections_static_no_filter_check
        CHECK (kind <> 'STATIC' OR filter_json IS NULL),
    CONSTRAINT question_collections_dynamic_filter_check
        CHECK (kind <> 'DYNAMIC' OR filter_json IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS question_collections_quiz_id_position_idx
    ON question_collections (quiz_id, position);

CREATE TABLE IF NOT EXISTS question_collection_members (
    id uuid PRIMARY KEY,
    collection_id uuid NOT NULL REFERENCES question_collections (id) ON DELETE CASCADE,
    question_id uuid NOT NULL REFERENCES questions (id) ON DELETE RESTRICT,
    position integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT question_collection_members_collection_question_unique
        UNIQUE (collection_id, question_id),
    CONSTRAINT question_collection_members_collection_position_unique
        UNIQUE (collection_id, position)
);

CREATE INDEX IF NOT EXISTS question_collection_members_collection_id_idx
    ON question_collection_members (collection_id, position);
