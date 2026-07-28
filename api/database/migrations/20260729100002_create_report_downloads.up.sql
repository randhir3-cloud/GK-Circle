-- EXAM-P8-T04: Report download tracking
CREATE TABLE report_downloads (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_id       uuid NOT NULL REFERENCES generated_reports(id) ON DELETE CASCADE,
  downloaded_by   varchar(255) NOT NULL,
  downloaded_at   timestamptz NOT NULL DEFAULT now(),
  ip_address      text,
  user_agent      text
);

CREATE INDEX idx_report_downloads_report ON report_downloads(report_id, downloaded_at DESC);
-- +migrate Up
