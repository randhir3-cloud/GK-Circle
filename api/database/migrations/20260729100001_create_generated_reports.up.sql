-- EXAM-P8-T04: Generated reports (export jobs) table
CREATE TABLE generated_reports (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  scheduled_report_id  uuid REFERENCES scheduled_reports(id) ON DELETE SET NULL,
  instructor_id        varchar(255) NOT NULL,
  title                text NOT NULL,
  export_type          text NOT NULL,
  export_format        text NOT NULL,
  status               text NOT NULL DEFAULT 'QUEUED',
  -- QUEUED | RUNNING | COMPLETED | FAILED | CANCELLED
  storage_key          text,            -- cleared on soft-delete or expiry
  storage_provider     text,
  file_size_bytes      bigint,
  row_count            integer,
  filters_json         jsonb NOT NULL DEFAULT '{}',
  quiz_id              uuid,
  error_message        text,
  snapshot_started_at  timestamptz,     -- REPEATABLE READ TX start; logical snapshot boundary
  queued_at            timestamptz NOT NULL DEFAULT now(),
  started_at           timestamptz,
  completed_at         timestamptz,
  expires_at           timestamptz,
  deleted_at           timestamptz,     -- soft-delete: storage purged, row retained
  created_at           timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_generated_reports_instructor ON generated_reports(instructor_id, created_at DESC);
CREATE INDEX idx_generated_reports_status     ON generated_reports(status)
  WHERE status IN ('QUEUED', 'RUNNING');
CREATE INDEX idx_generated_reports_expires    ON generated_reports(expires_at)
  WHERE status = 'COMPLETED' AND deleted_at IS NULL;
CREATE INDEX idx_generated_reports_reclaim    ON generated_reports(queued_at)
  WHERE status = 'QUEUED';
-- +migrate Up
