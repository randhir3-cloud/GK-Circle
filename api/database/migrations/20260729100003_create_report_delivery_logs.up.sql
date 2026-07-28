-- EXAM-P8-T04: Email delivery log with retry tracking
CREATE TABLE report_delivery_logs (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id        uuid NOT NULL REFERENCES generated_reports(id) ON DELETE CASCADE,
  recipient_email  text NOT NULL,
  delivered_at     timestamptz NOT NULL DEFAULT now(),
  delivery_status  text NOT NULL, -- SENT | FAILED | LINK_SENT | RETRYING
  attempt_number   integer NOT NULL DEFAULT 1,
  next_retry_at    timestamptz,
  error_message    text           -- SMTP error code/class only; no body, no URLs
);

CREATE INDEX idx_report_delivery_logs_report ON report_delivery_logs(report_id);
CREATE INDEX idx_report_delivery_logs_retry  ON report_delivery_logs(next_retry_at)
  WHERE delivery_status = 'RETRYING';
-- +migrate Up
