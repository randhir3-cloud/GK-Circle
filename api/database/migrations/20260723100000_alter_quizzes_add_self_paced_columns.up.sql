-- +migrate Up

-- Add self-paced assessment columns to the existing quizzes table.
-- Existing rows default to assessment_mode='LIVE' and status='DRAFT'.
-- The LIVE quiz code path (active_quizzes, socket controller) never reads
-- quizzes.status, so setting status='DRAFT' on existing rows is safe and
-- requires no backfill. Only SELF_PACED endpoints enforce status.

ALTER TABLE quizzes
  ADD COLUMN IF NOT EXISTS assessment_mode              varchar(20)   NOT NULL DEFAULT 'LIVE',
  ADD COLUMN IF NOT EXISTS status                       varchar(20)   NOT NULL DEFAULT 'DRAFT',
  ADD COLUMN IF NOT EXISTS published_at                 timestamp,
  ADD COLUMN IF NOT EXISTS duration_seconds             integer,
  ADD COLUMN IF NOT EXISTS max_attempts                 integer       NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS access_code                  varchar(16),
  ADD COLUMN IF NOT EXISTS negative_marks_per_question  numeric(4,2)  NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS show_leaderboard             boolean       NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS allow_answer_review          boolean       NOT NULL DEFAULT false;

ALTER TABLE quizzes
  ADD CONSTRAINT quizzes_assessment_mode_check
    CHECK (assessment_mode IN ('LIVE', 'SELF_PACED')),
  ADD CONSTRAINT quizzes_status_check
    CHECK (status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED'));

-- Unique index on access_code (sparse — only for non-NULL rows).
CREATE UNIQUE INDEX quizzes_access_code_uidx
  ON quizzes (access_code)
  WHERE access_code IS NOT NULL;
