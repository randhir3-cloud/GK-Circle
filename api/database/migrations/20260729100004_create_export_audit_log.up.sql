-- EXAM-P8-T04: Export audit log (metadata only — no personal data, no URLs)
CREATE TABLE export_audit_log (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  instructor_id   varchar(255) NOT NULL,
  report_id       uuid,
  action          text NOT NULL,
  -- EXPORT_REQUESTED | EXPORT_COMPLETED | EXPORT_FAILED
  -- EXPORT_DOWNLOADED | EXPORT_DELETED  | EXPORT_EXPIRED
  -- SCHEDULE_CREATED  | SCHEDULE_UPDATED | SCHEDULE_DELETED
  -- EMAIL_SENT        | EMAIL_FAILED     | EMAIL_RETRIED
  export_type     text,
  export_format   text,
  filters_json    jsonb, -- query params only; NO learner data, URLs, or signed tokens
  duration_ms     integer,
  row_count       integer,
  success         boolean,
  occurred_at     timestamptz NOT NULL DEFAULT now(),
  correlation_id  text
);

CREATE INDEX idx_export_audit_log_instructor ON export_audit_log(instructor_id, occurred_at DESC);
CREATE INDEX idx_export_audit_log_report     ON export_audit_log(report_id)
  WHERE report_id IS NOT NULL;
-- +migrate Up
