-- +migrate Down
-- Restore CASCADE delete from assessment_attempts to attempt_answers.

ALTER TABLE attempt_answers
    DROP CONSTRAINT IF EXISTS attempt_answers_attempt_fk;

ALTER TABLE attempt_answers
    ADD CONSTRAINT attempt_answers_attempt_fk
        FOREIGN KEY (attempt_id) REFERENCES assessment_attempts (id) ON DELETE CASCADE;
