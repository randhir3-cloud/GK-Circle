-- +migrate Up
-- EXAM-P3-T01: CSV import jobs store validation preview without persisting questions.

CREATE TABLE IF NOT EXISTS question_import_jobs (
    id uuid PRIMARY KEY,
    quiz_id uuid NOT NULL REFERENCES quizzes (id) ON DELETE CASCADE,
    created_by varchar(36) NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'PREVIEWED',
    source_filename text NOT NULL DEFAULT '',
    total_rows integer NOT NULL DEFAULT 0,
    valid_row_count integer NOT NULL DEFAULT 0,
    error_row_count integer NOT NULL DEFAULT 0,
    preview_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT question_import_jobs_status_check
        CHECK (status IN ('PREVIEWED', 'FAILED'))
);

CREATE INDEX IF NOT EXISTS question_import_jobs_quiz_id_created_at_idx
    ON question_import_jobs (quiz_id, created_at DESC);
