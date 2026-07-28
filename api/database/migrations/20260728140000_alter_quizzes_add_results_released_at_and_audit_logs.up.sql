-- +migrate Up

ALTER TABLE quizzes
  ADD COLUMN IF NOT EXISTS results_released_at timestamptz;

CREATE TABLE IF NOT EXISTS quiz_result_release_audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
  actor_id varchar(255) NOT NULL,
  event_type varchar(50) NOT NULL,
  previous_policy varchar(20),
  new_policy varchar(20),
  previous_state jsonb,
  new_state jsonb,
  ip_address varchar(45),
  user_agent text,
  schema_version smallint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT NOW()
);
