-- +migrate Up
-- EXAM-P5-T01 / ADR-024 §7: retain attempt_answers when parent attempt is deleted.

ALTER TABLE attempt_answers
    DROP CONSTRAINT IF EXISTS attempt_answers_attempt_fk;

ALTER TABLE attempt_answers
    ADD CONSTRAINT attempt_answers_attempt_fk
        FOREIGN KEY (attempt_id) REFERENCES assessment_attempts (id) ON DELETE RESTRICT;
