-- +migrate Down

DROP INDEX IF EXISTS question_import_jobs_quiz_id_created_at_idx;
DROP TABLE IF EXISTS question_import_jobs;
