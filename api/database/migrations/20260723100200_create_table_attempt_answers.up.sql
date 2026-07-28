-- +migrate Up

-- Per-question answer rows for self-paced attempts.
-- Rows are created on first save (not pre-created at attempt start).
-- score and is_correct are NULL until the attempt is submitted/auto-submitted
-- and the scoring service writes them in a single transaction.
-- is_marked_review allows the student to flag questions for re-visit.

CREATE TABLE attempt_answers (
  id                  uuid         NOT NULL DEFAULT gen_random_uuid(),
  attempt_id          uuid         NOT NULL,
  question_id         uuid         NOT NULL,
  selected_options    jsonb,
  is_marked_review    boolean      NOT NULL DEFAULT false,
  answered_at         timestamp,
  time_taken_seconds  integer,
  score               numeric(6,2),
  is_correct          boolean,
  created_at          timestamp    NOT NULL DEFAULT now(),
  updated_at          timestamp    NOT NULL DEFAULT now(),

  CONSTRAINT attempt_answers_pkey
    PRIMARY KEY (id),
  CONSTRAINT attempt_answers_attempt_fk
    FOREIGN KEY (attempt_id) REFERENCES assessment_attempts(id) ON DELETE CASCADE,
  CONSTRAINT attempt_answers_question_fk
    FOREIGN KEY (question_id) REFERENCES questions(id),
  CONSTRAINT attempt_answers_unique_per_question
    UNIQUE (attempt_id, question_id)
);

CREATE INDEX attempt_answers_attempt_idx
  ON attempt_answers (attempt_id);
