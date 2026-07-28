-- EXAM-P8-T04: Scheduled reports table
CREATE TABLE scheduled_reports (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  instructor_id    varchar(255) NOT NULL,
  title            text NOT NULL,
  export_type      text NOT NULL,
  -- PORTFOLIO_OVERVIEW | QUIZ_LIST | LEARNER_PERFORMANCE | RELEASE_MONITORING
  -- TIMELINE | QUIZ_SUMMARY | QUIZ_ATTEMPTS | QUESTION_METRICS | ENGAGEMENT | FULL_DASHBOARD
  export_format    text NOT NULL, -- CSV | XLSX | PDF
  schedule_type    text NOT NULL, -- DAILY | WEEKLY | MONTHLY | ONE_TIME
  cron_expr        text,          -- NULL for ONE_TIME; validated at creation (400 on invalid)
  timezone         text NOT NULL DEFAULT 'UTC', -- IANA timezone; validated against supported set
  next_run_at      timestamptz,
  last_run_at      timestamptz,
  enabled          boolean NOT NULL DEFAULT true,
  filters_json     jsonb NOT NULL DEFAULT '{}',
  quiz_id          uuid,          -- NULL for portfolio-scope reports
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_scheduled_reports_instructor ON scheduled_reports(instructor_id, enabled);
CREATE INDEX idx_scheduled_reports_next_run   ON scheduled_reports(next_run_at) WHERE enabled = true;
-- +migrate Up
