-- +migrate Up
-- EXAM-P2-T01: question versioning foundation and answer authority fields.

ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS lineage_id uuid,
    ADD COLUMN IF NOT EXISTS revision_number integer NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS official_answer json,
    ADD COLUMN IF NOT EXISTS authoritative_answer json,
    ADD COLUMN IF NOT EXISTS answer_review_status varchar(20) NOT NULL DEFAULT 'UNREVIEWED',
    ADD COLUMN IF NOT EXISTS answer_revision_reason text,
    ADD COLUMN IF NOT EXISTS answer_revision_source text;

UPDATE questions
SET lineage_id = id
WHERE lineage_id IS NULL;

UPDATE questions
SET official_answer = answers
WHERE official_answer IS NULL;

UPDATE questions
SET authoritative_answer = answers
WHERE authoritative_answer IS NULL;

UPDATE questions
SET answer_review_status = 'CONFIRMED'
WHERE answer_review_status = 'UNREVIEWED';

ALTER TABLE questions
    ALTER COLUMN lineage_id SET NOT NULL;

CREATE TABLE IF NOT EXISTS question_revisions (
    id uuid PRIMARY KEY,
    question_id uuid NOT NULL REFERENCES questions (id) ON DELETE RESTRICT,
    lineage_id uuid NOT NULL,
    revision_number integer NOT NULL,
    question text NOT NULL,
    type integer NOT NULL,
    options json NOT NULL,
    answers json NOT NULL,
    official_answer json NOT NULL,
    authoritative_answer json NOT NULL,
    answer_review_status varchar(20) NOT NULL,
    answer_revision_reason text,
    answer_revision_source text,
    points smallint,
    duration_in_seconds integer,
    question_media varchar(10),
    options_media varchar(10),
    resource text,
    created_by varchar(255),
    created_at timestamp NOT NULL DEFAULT (now()),
    CONSTRAINT question_revisions_lineage_revision_unique UNIQUE (lineage_id, revision_number)
);

CREATE INDEX question_revisions_lineage_id_idx ON question_revisions (lineage_id);
CREATE INDEX question_revisions_question_id_idx ON question_revisions (question_id);

INSERT INTO question_revisions (
    id,
    question_id,
    lineage_id,
    revision_number,
    question,
    type,
    options,
    answers,
    official_answer,
    authoritative_answer,
    answer_review_status,
    answer_revision_reason,
    answer_revision_source,
    points,
    duration_in_seconds,
    question_media,
    options_media,
    resource,
    created_at
)
SELECT
    gen_random_uuid(),
    q.id,
    q.lineage_id,
    q.revision_number,
    q.question,
    q.type,
    q.options,
    q.answers,
    q.official_answer,
    q.authoritative_answer,
    q.answer_review_status,
    q.answer_revision_reason,
    q.answer_revision_source,
    q.points,
    q.duration_in_seconds,
    q.question_media,
    q.options_media,
    q.resource,
    q.created_at
FROM questions q
WHERE NOT EXISTS (
    SELECT 1
    FROM question_revisions qr
    WHERE qr.lineage_id = q.lineage_id
      AND qr.revision_number = q.revision_number
);
