-- +migrate Up

-- Self-paced attempt tracking.
-- One row per user per attempt number per quiz.
-- question_order stores the ordered list of question UUIDs set at creation;
-- it never changes after creation to ensure a stable question sequence on resume.
-- status follows a strict state machine:
--   IN_PROGRESS → SUBMITTED (manual) | AUTO_SUBMITTED (timer expiry) | ABANDONED
-- EXPIRED is intentionally absent: an expired attempt is always AUTO_SUBMITTED.

CREATE TABLE assessment_attempts (
  id                  uuid         NOT NULL DEFAULT gen_random_uuid(),
  quiz_id             uuid         NOT NULL,
  user_id             varchar(20)  NOT NULL,
  attempt_number      integer      NOT NULL DEFAULT 1,
  status              varchar(20)  NOT NULL DEFAULT 'IN_PROGRESS',
  question_order      jsonb        NOT NULL DEFAULT '[]',
  started_at          timestamp    NOT NULL DEFAULT now(),
  submitted_at        timestamp,
  expires_at          timestamp,
  total_score         numeric(8,2),
  max_score           numeric(8,2),
  time_taken_seconds  integer,
  created_at          timestamp    NOT NULL DEFAULT now(),
  updated_at          timestamp    NOT NULL DEFAULT now(),

  CONSTRAINT assessment_attempts_pkey
    PRIMARY KEY (id),
  CONSTRAINT assessment_attempts_quiz_fk
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id),
  CONSTRAINT assessment_attempts_user_fk
    FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT assessment_attempts_status_check
    CHECK (status IN ('IN_PROGRESS', 'SUBMITTED', 'AUTO_SUBMITTED', 'ABANDONED')),
  CONSTRAINT assessment_attempts_unique_per_attempt
    UNIQUE (quiz_id, user_id, attempt_number)
);

CREATE INDEX assessment_attempts_quiz_status_idx
  ON assessment_attempts (quiz_id, status);

CREATE INDEX assessment_attempts_user_idx
  ON assessment_attempts (user_id, status);
