-- +migrate Down

DROP TABLE IF EXISTS question_revisions;

ALTER TABLE questions
    DROP COLUMN IF EXISTS lineage_id,
    DROP COLUMN IF EXISTS revision_number,
    DROP COLUMN IF EXISTS official_answer,
    DROP COLUMN IF EXISTS authoritative_answer,
    DROP COLUMN IF EXISTS answer_review_status,
    DROP COLUMN IF EXISTS answer_revision_reason,
    DROP COLUMN IF EXISTS answer_revision_source;
