-- +migrate Up

-- Add result release policies and independent review controls to quizzes table.
ALTER TABLE quizzes
  ADD COLUMN IF NOT EXISTS result_release_policy varchar(20) NOT NULL DEFAULT 'IMMEDIATE',
  ADD COLUMN IF NOT EXISTS results_released       boolean     NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS results_scheduled_at  timestamptz,
  ADD COLUMN IF NOT EXISTS show_score             boolean     NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS show_pass_fail        boolean     NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS show_correctness       boolean     NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS show_explanations      boolean     NOT NULL DEFAULT true;

ALTER TABLE quizzes
  DROP CONSTRAINT IF EXISTS quizzes_result_release_policy_check;

ALTER TABLE quizzes
  ADD CONSTRAINT quizzes_result_release_policy_check
    CHECK (result_release_policy IN ('IMMEDIATE', 'MANUAL', 'SCHEDULED'));
