-- +migrate Up
-- +migrate StatementBegin
DO $$
DECLARE
    t_exists BOOLEAN;
    c_exists BOOLEAN;
    c_type TEXT;
    c_nullable TEXT;
    c_default TEXT;
BEGIN
    ---------------------------------------------------------------------------
    -- 1. Table: public.scheduled_reports
    ---------------------------------------------------------------------------
    SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'scheduled_reports') INTO t_exists;
    IF t_exists THEN
        -- instructor_id
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'instructor_id') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.scheduled_reports is missing column: instructor_id'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'instructor_id';
        IF c_type NOT IN ('character varying', 'text') THEN RAISE EXCEPTION 'Table public.scheduled_reports column instructor_id has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.scheduled_reports column instructor_id must be NOT NULL'; END IF;

        -- id
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'id') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.scheduled_reports is missing column: id'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'id';
        IF c_type <> 'uuid' THEN RAISE EXCEPTION 'Table public.scheduled_reports column id has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.scheduled_reports column id must be NOT NULL'; END IF;

        -- title
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'title') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.scheduled_reports is missing column: title'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'title';
        IF c_type <> 'text' THEN RAISE EXCEPTION 'Table public.scheduled_reports column title has incompatible type: %', c_type; END IF;

        -- next_run_at
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'next_run_at') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.scheduled_reports is missing column: next_run_at'; END IF;
        SELECT data_type INTO c_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'next_run_at';
        IF c_type <> 'timestamp with time zone' THEN RAISE EXCEPTION 'Table public.scheduled_reports column next_run_at has incompatible type: %', c_type; END IF;

        -- enabled
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'enabled') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.scheduled_reports is missing column: enabled'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'enabled';
        IF c_type <> 'boolean' THEN RAISE EXCEPTION 'Table public.scheduled_reports column enabled has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.scheduled_reports column enabled must be NOT NULL'; END IF;

        -- timezone
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'timezone') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.scheduled_reports is missing column: timezone'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'scheduled_reports' AND column_name = 'timezone';
        IF c_type <> 'text' THEN RAISE EXCEPTION 'Table public.scheduled_reports column timezone has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.scheduled_reports column timezone must be NOT NULL'; END IF;

    ELSE
        CREATE TABLE public.scheduled_reports (
            id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            instructor_id    varchar(255) NOT NULL,
            title            text NOT NULL,
            export_type      text NOT NULL,
            export_format    text NOT NULL,
            schedule_type    text NOT NULL,
            cron_expr        text,
            timezone         text NOT NULL DEFAULT 'UTC',
            next_run_at      timestamptz,
            last_run_at      timestamptz,
            enabled          boolean NOT NULL DEFAULT true,
            filters_json     jsonb NOT NULL DEFAULT '{}',
            quiz_id          uuid,
            created_at       timestamptz NOT NULL DEFAULT now(),
            updated_at       timestamptz NOT NULL DEFAULT now()
        );
    END IF;

    CREATE INDEX IF NOT EXISTS idx_scheduled_reports_instructor ON public.scheduled_reports(instructor_id, enabled);
    CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run   ON public.scheduled_reports(next_run_at) WHERE enabled = true;

    ---------------------------------------------------------------------------
    -- 2. Table: public.generated_reports
    ---------------------------------------------------------------------------
    SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'generated_reports') INTO t_exists;
    IF t_exists THEN
        -- scheduled_report_id
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'generated_reports' AND column_name = 'scheduled_report_id') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.generated_reports is missing column: scheduled_report_id'; END IF;
        SELECT data_type INTO c_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'generated_reports' AND column_name = 'scheduled_report_id';
        IF c_type <> 'uuid' THEN RAISE EXCEPTION 'Table public.generated_reports column scheduled_report_id has incompatible type: %', c_type; END IF;

        -- status
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'generated_reports' AND column_name = 'status') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.generated_reports is missing column: status'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'generated_reports' AND column_name = 'status';
        IF c_type <> 'text' THEN RAISE EXCEPTION 'Table public.generated_reports column status has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.generated_reports column status must be NOT NULL'; END IF;

    ELSE
        CREATE TABLE public.generated_reports (
            id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            scheduled_report_id  uuid REFERENCES public.scheduled_reports(id) ON DELETE SET NULL,
            instructor_id        varchar(255) NOT NULL,
            title                text NOT NULL,
            export_type          text NOT NULL,
            export_format        text NOT NULL,
            status               text NOT NULL DEFAULT 'QUEUED',
            storage_key          text,
            storage_provider     text,
            file_size_bytes      bigint,
            row_count            integer,
            filters_json         jsonb NOT NULL DEFAULT '{}',
            quiz_id              uuid,
            error_message        text,
            snapshot_started_at  timestamptz,
            queued_at            timestamptz NOT NULL DEFAULT now(),
            started_at           timestamptz,
            completed_at         timestamptz,
            expires_at           timestamptz,
            deleted_at           timestamptz,
            created_at           timestamptz NOT NULL DEFAULT now()
        );
    END IF;

    CREATE INDEX IF NOT EXISTS idx_generated_reports_instructor ON public.generated_reports(instructor_id, created_at DESC);
    CREATE INDEX IF NOT EXISTS idx_generated_reports_status     ON public.generated_reports(status) WHERE status IN ('QUEUED', 'RUNNING');
    CREATE INDEX IF NOT EXISTS idx_generated_reports_expires    ON public.generated_reports(expires_at) WHERE status = 'COMPLETED' AND deleted_at IS NULL;
    CREATE INDEX IF NOT EXISTS idx_generated_reports_reclaim    ON public.generated_reports(queued_at) WHERE status = 'QUEUED';

    ---------------------------------------------------------------------------
    -- 3. Table: public.report_downloads
    ---------------------------------------------------------------------------
    SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'report_downloads') INTO t_exists;
    IF t_exists THEN
        -- report_id
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'report_downloads' AND column_name = 'report_id') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.report_downloads is missing column: report_id'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'report_downloads' AND column_name = 'report_id';
        IF c_type <> 'uuid' THEN RAISE EXCEPTION 'Table public.report_downloads column report_id has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.report_downloads column report_id must be NOT NULL'; END IF;

    ELSE
        CREATE TABLE public.report_downloads (
            id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            report_id       uuid NOT NULL REFERENCES public.generated_reports(id) ON DELETE CASCADE,
            downloaded_by   varchar(255) NOT NULL,
            downloaded_at   timestamptz NOT NULL DEFAULT now(),
            ip_address      text,
            user_agent      text
        );
    END IF;

    CREATE INDEX IF NOT EXISTS idx_report_downloads_report ON public.report_downloads(report_id, downloaded_at DESC);

    ---------------------------------------------------------------------------
    -- 4. Table: public.report_delivery_logs
    ---------------------------------------------------------------------------
    SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'report_delivery_logs') INTO t_exists;
    IF t_exists THEN
        -- report_id
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'report_delivery_logs' AND column_name = 'report_id') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.report_delivery_logs is missing column: report_id'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'report_delivery_logs' AND column_name = 'report_id';
        IF c_type <> 'uuid' THEN RAISE EXCEPTION 'Table public.report_delivery_logs column report_id has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.report_delivery_logs column report_id must be NOT NULL'; END IF;

    ELSE
        CREATE TABLE public.report_delivery_logs (
            id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            report_id        uuid NOT NULL REFERENCES public.generated_reports(id) ON DELETE CASCADE,
            recipient_email  text NOT NULL,
            delivered_at     timestamptz NOT NULL DEFAULT now(),
            delivery_status  text NOT NULL,
            attempt_number   integer NOT NULL DEFAULT 1,
            next_retry_at    timestamptz,
            error_message    text
        );
    END IF;

    CREATE INDEX IF NOT EXISTS idx_report_delivery_logs_report ON public.report_delivery_logs(report_id);
    CREATE INDEX IF NOT EXISTS idx_report_delivery_logs_retry  ON public.report_delivery_logs(next_retry_at) WHERE delivery_status = 'RETRYING';

    ---------------------------------------------------------------------------
    -- 5. Table: public.export_audit_log
    ---------------------------------------------------------------------------
    SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'export_audit_log') INTO t_exists;
    IF t_exists THEN
        -- action
        SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'export_audit_log' AND column_name = 'action') INTO c_exists;
        IF NOT c_exists THEN RAISE EXCEPTION 'Table public.export_audit_log is missing column: action'; END IF;
        SELECT data_type, is_nullable INTO c_type, c_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'export_audit_log' AND column_name = 'action';
        IF c_type <> 'text' THEN RAISE EXCEPTION 'Table public.export_audit_log column action has incompatible type: %', c_type; END IF;
        IF c_nullable <> 'NO' THEN RAISE EXCEPTION 'Table public.export_audit_log column action must be NOT NULL'; END IF;

    ELSE
        CREATE TABLE public.export_audit_log (
            id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            instructor_id   varchar(255) NOT NULL,
            report_id       uuid,
            action          text NOT NULL,
            export_type     text,
            export_format   text,
            filters_json    jsonb,
            duration_ms     integer,
            row_count       integer,
            success         boolean,
            occurred_at     timestamptz NOT NULL DEFAULT now(),
            correlation_id  text
        );
    END IF;

    CREATE INDEX IF NOT EXISTS idx_export_audit_log_instructor ON public.export_audit_log(instructor_id, occurred_at DESC);
    CREATE INDEX IF NOT EXISTS idx_export_audit_log_report     ON public.export_audit_log(report_id) WHERE report_id IS NOT NULL;

END $$;
-- +migrate StatementEnd

-- +migrate Down
-- Intentionally irreversible to protect production report records.
SELECT 1;
