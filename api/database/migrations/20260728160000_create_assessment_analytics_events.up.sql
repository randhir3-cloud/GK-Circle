-- +migrate Up

CREATE TABLE IF NOT EXISTS assessment_analytics_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  client_event_id uuid,

  -- Immutable snapshots (survive entity deletion)
  attempt_id uuid NOT NULL,
  quiz_id uuid NOT NULL,
  user_id varchar(255) NOT NULL,
  quiz_owner_id varchar(255),

  -- Nullable FK references
  attempt_ref_id uuid REFERENCES assessment_attempts(id) ON DELETE SET NULL,
  quiz_ref_id uuid REFERENCES quizzes(id) ON DELETE SET NULL,

  event_type varchar(50) NOT NULL,
  event_source varchar(20) NOT NULL,
  correlation_id varchar(255) NOT NULL,
  idempotency_key varchar(255),

  schema_version smallint NOT NULL DEFAULT 1,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_analytics_metadata_object CHECK (jsonb_typeof(metadata) = 'object'),
  CONSTRAINT chk_analytics_event_source CHECK (
    event_source IN ('HTTP', 'WORKER', 'SCHEDULER', 'CLIENT_BATCH')
  ),
  CONSTRAINT chk_analytics_event_type CHECK (
    event_type IN (
      'ATTEMPT_STARTED',
      'ATTEMPT_SUBMITTED',
      'ATTEMPT_AUTO_SUBMITTED',
      'RESULT_RELEASE_OVERRIDE_APPLIED',
      'RESULT_RELEASE_SCHEDULED_EFFECTIVE',
      'ATTEMPT_AUTOSAVED',
      'RESULT_VIEWED',
      'QUESTION_VIEWED',
      'ANSWER_SELECTED',
      'ANSWER_CHANGED',
      'QUESTION_TIME_SPENT',
      'HINT_OPENED',
      'REVIEW_OPENED'
    )
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_client_event
  ON assessment_analytics_events(client_event_id)
  WHERE client_event_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_analytics_attempt_idempotency
  ON assessment_analytics_events(user_id, attempt_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_analytics_attempt_created
  ON assessment_analytics_events(attempt_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_quiz_created
  ON assessment_analytics_events(quiz_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_user_created
  ON assessment_analytics_events(user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_correlation_id
  ON assessment_analytics_events(correlation_id);
