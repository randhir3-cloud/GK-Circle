-- +migrate Down

ALTER TABLE question_import_jobs
    DROP COLUMN IF EXISTS committed_at,
    DROP COLUMN IF EXISTS commit_result_json;

ALTER TABLE question_import_jobs
    DROP CONSTRAINT IF EXISTS question_import_jobs_status_check;

ALTER TABLE question_import_jobs
    ADD CONSTRAINT question_import_jobs_status_check
        CHECK (status IN ('PREVIEWED', 'FAILED'));
