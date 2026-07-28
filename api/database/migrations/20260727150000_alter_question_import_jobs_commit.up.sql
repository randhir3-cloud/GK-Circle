-- +migrate Up
-- EXAM-P3-T02: import job commit workflow (status transitions + commit result).

ALTER TABLE question_import_jobs
    DROP CONSTRAINT IF EXISTS question_import_jobs_status_check;

ALTER TABLE question_import_jobs
    ADD CONSTRAINT question_import_jobs_status_check
        CHECK (status IN ('PREVIEWED', 'FAILED', 'COMMITTING', 'COMMITTED'));

ALTER TABLE question_import_jobs
    ADD COLUMN IF NOT EXISTS commit_result_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS committed_at timestamptz;
