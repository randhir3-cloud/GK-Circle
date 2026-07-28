package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goqulib "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

// Sentinel errors for the export service.
var (
	ErrReportNotFound    = errors.New("report not found")
	ErrReportUnauthorized = errors.New("report access denied")
	ErrReportDeleted     = errors.New("report has been deleted")
	ErrReportNotCompleted = errors.New("report is not completed")
	ErrScheduleNotFound  = errors.New("schedule not found")
)

// DownloadMeta is the minimal metadata required to serve a download.
type DownloadMeta struct {
	StorageKey   string
	ExportFormat string
}

// ExportServiceDB wraps DB access for the export management layer.
// Separate from ExportService to keep generation logic isolated.
type ExportServiceDB struct {
	db     *goqulib.Database
	logger *zap.Logger
}

// NewExportServiceDB creates the DB management layer.
func NewExportServiceDB(db *goqulib.Database, logger *zap.Logger) *ExportServiceDB {
	return &ExportServiceDB{db: db, logger: logger}
}

// InsertJob inserts a QUEUED generated_reports row.
func (s *ExportService) InsertJob(
	ctx context.Context,
	jobID uuid.UUID,
	instructorID, title string,
	exportType structs.ExportType,
	exportFormat structs.ExportFormat,
	filtersJSON json.RawMessage,
	quizIDStr *string,
) error {
	var quizID interface{} = nil
	if quizIDStr != nil {
		quizID = *quizIDStr
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO generated_reports
			(id, instructor_id, title, export_type, export_format, status, filters_json, quiz_id)
		VALUES ($1, $2, $3, $4, $5, 'QUEUED', $6, $7)`,
		jobID, instructorID, title, string(exportType), string(exportFormat), []byte(filtersJSON), quizID,
	)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	// Audit: export requested.
	s.writeAudit(ctx, instructorID, &jobID, "EXPORT_REQUESTED", string(exportType), string(exportFormat), filtersJSON, 0, 0, true)
	return nil
}

// GetJob loads a job and verifies ownership.
func (s *ExportService) GetJob(ctx context.Context, reportID uuid.UUID, instructorID string) (*structs.ExportJobResponse, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, status, export_type, export_format, title, quiz_id,
		       row_count, file_size_bytes, error_message, snapshot_started_at,
		       queued_at, started_at, completed_at, expires_at, deleted_at,
		       scheduled_report_id, instructor_id
		FROM generated_reports
		WHERE id = $1
	`, reportID)

	var r struct {
		id, status, exportType, exportFormat, title string
		quizID, errorMsg, scheduledReportID         sql.NullString
		ownerID                                      string
		rowCount                                     sql.NullInt64
		fileSizeBytes                                sql.NullInt64
		snapshotAt                                   sql.NullTime
		queuedAt                                     time.Time
		startedAt, completedAt, expiresAt, deletedAt sql.NullTime
	}
	if err := row.Scan(
		&r.id, &r.status, &r.exportType, &r.exportFormat, &r.title,
		&r.quizID, &r.rowCount, &r.fileSizeBytes, &r.errorMsg, &r.snapshotAt,
		&r.queuedAt, &r.startedAt, &r.completedAt, &r.expiresAt, &r.deletedAt,
		&r.scheduledReportID, &r.ownerID,
	); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReportNotFound
	} else if err != nil {
		return nil, err
	}

	if r.ownerID != instructorID {
		return nil, ErrReportUnauthorized
	}

	resp := &structs.ExportJobResponse{
		ID:           r.id,
		Status:       structs.ReportStatus(r.status),
		ExportType:   structs.ExportType(r.exportType),
		ExportFormat: structs.ExportFormat(r.exportFormat),
		Title:        r.title,
		QueuedAt:     r.queuedAt,
	}
	if r.quizID.Valid {
		resp.QuizID = &r.quizID.String
	}
	if r.rowCount.Valid {
		n := int(r.rowCount.Int64)
		resp.RowCount = &n
	}
	if r.fileSizeBytes.Valid {
		resp.FileSizeBytes = &r.fileSizeBytes.Int64
	}
	if r.errorMsg.Valid {
		resp.ErrorMessage = &r.errorMsg.String
	}
	if r.snapshotAt.Valid {
		resp.SnapshotStartedAt = &r.snapshotAt.Time
	}
	if r.startedAt.Valid {
		resp.StartedAt = &r.startedAt.Time
	}
	if r.completedAt.Valid {
		resp.CompletedAt = &r.completedAt.Time
	}
	if r.expiresAt.Valid {
		resp.ExpiresAt = &r.expiresAt.Time
	}
	if r.deletedAt.Valid {
		resp.DeletedAt = &r.deletedAt.Time
	}
	if r.scheduledReportID.Valid {
		resp.ScheduledReportID = &r.scheduledReportID.String
	}
	return resp, nil
}

// GetDownloadMeta loads the metadata needed to stream a download. Verifies ownership.
func (s *ExportService) GetDownloadMeta(ctx context.Context, reportID uuid.UUID, instructorID string) (*DownloadMeta, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT instructor_id, status, storage_key, export_format, deleted_at
		 FROM generated_reports WHERE id = $1`, reportID)

	var ownerID, status, exportFormat string
	var storageKey, deletedAt sql.NullString
	var deletedAtTime sql.NullTime

	if err := row.Scan(&ownerID, &status, &storageKey, &exportFormat, &deletedAtTime); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReportNotFound
	} else if err != nil {
		return nil, err
	}
	_ = deletedAt

	if ownerID != instructorID {
		return nil, ErrReportUnauthorized
	}
	if deletedAtTime.Valid {
		return nil, ErrReportDeleted
	}
	if status != string(structs.ReportStatusCompleted) {
		return nil, ErrReportNotCompleted
	}
	if !storageKey.Valid || storageKey.String == "" {
		return nil, ErrReportDeleted
	}
	return &DownloadMeta{StorageKey: storageKey.String, ExportFormat: exportFormat}, nil
}

// RecordDownload inserts a report_downloads row.
func (s *ExportService) RecordDownload(ctx context.Context, reportID uuid.UUID, downloadedBy, ipAddress, userAgent string) {
	s.db.ExecContext(ctx,
		`INSERT INTO report_downloads (report_id, downloaded_by, ip_address, user_agent) VALUES ($1, $2, $3, $4)`,
		reportID, downloadedBy, ipAddress, truncateStr(userAgent, 512),
	)
	s.writeAudit(ctx, downloadedBy, &reportID, "EXPORT_DOWNLOADED", "", "", nil, 0, 0, true)
}

// DeleteReport soft-deletes COMPLETED reports (purge storage, set deleted_at) or cancels queued ones.
func (s *ExportService) DeleteReport(ctx context.Context, reportID uuid.UUID, instructorID string, storage StorageProvider) error {
	row := s.db.QueryRowContext(ctx,
		`SELECT instructor_id, status, storage_key FROM generated_reports WHERE id = $1`, reportID)

	var ownerID, status string
	var storageKey sql.NullString
	if err := row.Scan(&ownerID, &status, &storageKey); errors.Is(err, sql.ErrNoRows) {
		return ErrReportNotFound
	} else if err != nil {
		return err
	}
	if ownerID != instructorID {
		return ErrReportUnauthorized
	}

	switch structs.ReportStatus(status) {
	case structs.ReportStatusQueued, structs.ReportStatusRunning:
		_, err := s.db.ExecContext(ctx,
			`UPDATE generated_reports SET status = 'CANCELLED' WHERE id = $1 AND status = $2`,
			reportID, status)
		return err

	case structs.ReportStatusCompleted:
		// Purge storage object.
		if storageKey.Valid && storageKey.String != "" && storage != nil {
			if err := storage.Delete(ctx, storageKey.String); err != nil {
				s.logger.Error("delete report: storage delete", zap.Error(err))
				// Continue — clear DB key.
			}
		}
		// Soft-delete: clear storage_key, set deleted_at, status unchanged.
		_, err := s.db.ExecContext(ctx,
			`UPDATE generated_reports SET storage_key = NULL, deleted_at = now() WHERE id = $1`,
			reportID)
		if err == nil {
			s.writeAudit(ctx, instructorID, &reportID, "EXPORT_DELETED", "", "", nil, 0, 0, true)
		}
		return err

	default:
		// FAILED / CANCELLED — nothing to purge, just mark deleted.
		_, err := s.db.ExecContext(ctx,
			`UPDATE generated_reports SET deleted_at = now() WHERE id = $1`, reportID)
		return err
	}
}

// ListHistory returns a cursor-paginated list of generated reports for an instructor.
func (s *ExportService) ListHistory(ctx context.Context, instructorID, cursor string, limit int) (*structs.GeneratedReportListResponse, error) {
	query := `
		SELECT id, status, export_type, export_format, title, quiz_id,
		       row_count, file_size_bytes, queued_at, completed_at, expires_at, deleted_at
		FROM generated_reports
		WHERE instructor_id = $1
	`
	args := []interface{}{instructorID}
	if cursor != "" {
		t, err := decodeCursorTime(cursor)
		if err == nil {
			query += fmt.Sprintf(` AND created_at < $%d`, len(args)+1)
			args = append(args, t)
		}
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []structs.GeneratedReportListItem
	for rows.Next() {
		var item structs.GeneratedReportListItem
		var quizID sql.NullString
		var rowCount, fileSize sql.NullInt64
		var completedAt, expiresAt, deletedAt, queuedAt sql.NullTime

		if err := rows.Scan(
			&item.ID, (*string)(&item.Status), (*string)(&item.ExportType), (*string)(&item.ExportFormat), &item.Title,
			&quizID, &rowCount, &fileSize, &queuedAt, &completedAt, &expiresAt, &deletedAt,
		); err != nil {
			return nil, err
		}
		if queuedAt.Valid {
			item.QueuedAt = queuedAt.Time
		}
		if quizID.Valid {
			item.QuizID = &quizID.String
		}
		if rowCount.Valid {
			n := int(rowCount.Int64)
			item.RowCount = &n
		}
		if fileSize.Valid {
			item.FileSizeBytes = &fileSize.Int64
		}
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		if deletedAt.Valid {
			item.DeletedAt = &deletedAt.Time
		}
		items = append(items, item)
	}

	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		cur := encodeCursorTime(items[len(items)-1].QueuedAt)
		nextCursor = &cur
	}
	return &structs.GeneratedReportListResponse{Items: items, NextCursor: nextCursor, Total: len(items)}, nil
}

// ListQuizHistory returns reports for a specific quiz.
func (s *ExportService) ListQuizHistory(ctx context.Context, instructorID, quizID, cursor string, limit int) (*structs.GeneratedReportListResponse, error) {
	query := `
		SELECT id, status, export_type, export_format, title, quiz_id,
		       row_count, file_size_bytes, queued_at, completed_at, expires_at, deleted_at
		FROM generated_reports
		WHERE instructor_id = $1 AND quiz_id = $2
	`
	args := []interface{}{instructorID, quizID}
	if cursor != "" {
		if t, err := decodeCursorTime(cursor); err == nil {
			query += fmt.Sprintf(` AND created_at < $%d`, len(args)+1)
			args = append(args, t)
		}
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []structs.GeneratedReportListItem
	for rows.Next() {
		var item structs.GeneratedReportListItem
		var quizIDNull sql.NullString
		var rowCount, fileSize sql.NullInt64
		var completedAt, expiresAt, deletedAt, queuedAt sql.NullTime
		if err := rows.Scan(
			&item.ID, (*string)(&item.Status), (*string)(&item.ExportType), (*string)(&item.ExportFormat), &item.Title,
			&quizIDNull, &rowCount, &fileSize, &queuedAt, &completedAt, &expiresAt, &deletedAt,
		); err != nil {
			return nil, err
		}
		if queuedAt.Valid {
			item.QueuedAt = queuedAt.Time
		}
		if quizIDNull.Valid {
			item.QuizID = &quizIDNull.String
		}
		if rowCount.Valid {
			n := int(rowCount.Int64)
			item.RowCount = &n
		}
		if fileSize.Valid {
			item.FileSizeBytes = &fileSize.Int64
		}
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}
		if expiresAt.Valid {
			item.ExpiresAt = &expiresAt.Time
		}
		if deletedAt.Valid {
			item.DeletedAt = &deletedAt.Time
		}
		items = append(items, item)
	}
	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		cur := encodeCursorTime(items[len(items)-1].QueuedAt)
		nextCursor = &cur
	}
	return &structs.GeneratedReportListResponse{Items: items, NextCursor: nextCursor, Total: len(items)}, nil
}

// ListAuditLog returns a cursor-paginated export audit log for an instructor.
func (s *ExportService) ListAuditLog(ctx context.Context, instructorID, cursor string, limit int) (*structs.AuditLogListResponse, error) {
	query := `
		SELECT id, action, export_type, export_format, report_id, filters_json,
		       duration_ms, row_count, success, occurred_at, correlation_id
		FROM export_audit_log
		WHERE instructor_id = $1
	`
	args := []interface{}{instructorID}
	if cursor != "" {
		if t, err := decodeCursorTime(cursor); err == nil {
			query += fmt.Sprintf(` AND occurred_at < $%d`, len(args)+1)
			args = append(args, t)
		}
	}
	query += fmt.Sprintf(` ORDER BY occurred_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []structs.AuditLogItem
	for rows.Next() {
		var item structs.AuditLogItem
		var exportType, exportFormat, reportID, corrID sql.NullString
		var filtersJSON []byte
		var durationMs, rowCount sql.NullInt64
		var success sql.NullBool
		var occurredAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Action, &exportType, &exportFormat, &reportID, &filtersJSON,
			&durationMs, &rowCount, &success, &occurredAt, &corrID,
		); err != nil {
			return nil, err
		}
		item.OccurredAt = occurredAt
		if exportType.Valid {
			item.ExportType = &exportType.String
		}
		if exportFormat.Valid {
			item.ExportFormat = &exportFormat.String
		}
		if reportID.Valid {
			item.ReportID = &reportID.String
		}
		if len(filtersJSON) > 0 {
			item.FiltersJSON = json.RawMessage(filtersJSON)
		}
		if durationMs.Valid {
			n := int(durationMs.Int64)
			item.DurationMs = &n
		}
		if rowCount.Valid {
			n := int(rowCount.Int64)
			item.RowCount = &n
		}
		if success.Valid {
			item.Success = &success.Bool
		}
		if corrID.Valid {
			item.CorrelationID = &corrID.String
		}
		items = append(items, item)
	}

	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		cur := encodeCursorTime(items[len(items)-1].OccurredAt)
		nextCursor = &cur
	}
	return &structs.AuditLogListResponse{Items: items, NextCursor: nextCursor}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Schedule CRUD
// ─────────────────────────────────────────────────────────────────────────────

func (s *ExportService) CreateSchedule(ctx context.Context, instructorID string, body structs.CreateScheduleInput) (*structs.ScheduledReportResponse, error) {
	id := uuid.New()
	filtersRaw, _ := json.Marshal(body.FiltersJSON)
	cronExpr := ""
	if body.CronExpr != nil {
		cronExpr = *body.CronExpr
	}

	// Compute initial next_run_at.
	var nextRun *time.Time
	if body.ScheduleType != structs.ScheduleTypeOneTime && cronExpr != "" {
		t, _ := computeNextRun(string(body.ScheduleType), cronExpr, body.Timezone)
		nextRun = t
	}

	var quizID interface{} = nil
	if body.QuizID != nil {
		quizID = *body.QuizID
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scheduled_reports
			(id, instructor_id, title, export_type, export_format, schedule_type, cron_expr, timezone,
			 next_run_at, enabled, filters_json, quiz_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, $10, $11)
	`, id, instructorID, body.Title, string(body.ExportType), string(body.ExportFormat),
		string(body.ScheduleType), nullableStr(cronExpr), body.Timezone, nextRun, filtersRaw, quizID)
	if err != nil {
		return nil, err
	}

	s.writeAudit(ctx, instructorID, nil, "SCHEDULE_CREATED", string(body.ExportType), string(body.ExportFormat), filtersRaw, 0, 0, true)

	return s.GetSchedule(ctx, id, instructorID)
}

func (s *ExportService) ListSchedules(ctx context.Context, instructorID string) (*structs.ScheduledReportListResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, export_type, export_format, schedule_type, cron_expr, timezone,
		       enabled, filters_json, quiz_id, next_run_at, last_run_at, created_at, updated_at
		FROM scheduled_reports
		WHERE instructor_id = $1
		ORDER BY created_at DESC
	`, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []structs.ScheduledReportResponse
	for rows.Next() {
		item, err := scanScheduleRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &structs.ScheduledReportListResponse{Items: items}, nil
}

func (s *ExportService) GetSchedule(ctx context.Context, id uuid.UUID, instructorID string) (*structs.ScheduledReportResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, export_type, export_format, schedule_type, cron_expr, timezone,
		       enabled, filters_json, quiz_id, next_run_at, last_run_at, created_at, updated_at
		FROM scheduled_reports
		WHERE id = $1 AND instructor_id = $2
	`, id, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrScheduleNotFound
	}
	item, err := scanScheduleRow(rows)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ExportService) UpdateSchedule(ctx context.Context, id uuid.UUID, instructorID string, body structs.UpdateScheduleInput) (*structs.ScheduledReportResponse, error) {
	if body.CronExpr != nil && *body.CronExpr != "" {
		if err := ValidateCronExpr(*body.CronExpr); err != nil {
			return nil, err
		}
	}

	setParts := []string{"updated_at = now()"}
	args := []interface{}{}
	argc := 1

	addArg := func(col string, v interface{}) {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", col, argc))
		args = append(args, v)
		argc++
	}

	if body.Title != nil {
		addArg("title", *body.Title)
	}
	if body.CronExpr != nil {
		addArg("cron_expr", *body.CronExpr)
	}
	if body.Timezone != nil {
		addArg("timezone", *body.Timezone)
	}
	if body.Enabled != nil {
		addArg("enabled", *body.Enabled)
	}
	if body.FiltersJSON != nil {
		raw, _ := json.Marshal(body.FiltersJSON)
		addArg("filters_json", raw)
	}

	set := ""
	for i, p := range setParts {
		if i > 0 {
			set += ", "
		}
		set += p
	}

	args = append(args, id, instructorID)
	query := fmt.Sprintf(
		`UPDATE scheduled_reports SET %s WHERE id = $%d AND instructor_id = $%d`,
		set, argc, argc+1,
	)
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrScheduleNotFound
	}
	s.writeAudit(ctx, instructorID, nil, "SCHEDULE_UPDATED", "", "", nil, 0, 0, true)
	return s.GetSchedule(ctx, id, instructorID)
}

func (s *ExportService) DeleteSchedule(ctx context.Context, id uuid.UUID, instructorID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM scheduled_reports WHERE id = $1 AND instructor_id = $2`, id, instructorID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrScheduleNotFound
	}
	s.writeAudit(ctx, instructorID, nil, "SCHEDULE_DELETED", "", "", nil, 0, 0, true)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanScheduleRow(row rowScanner) (structs.ScheduledReportResponse, error) {
	var id, title, exportType, exportFormat, scheduleType, timezone string
	var cronExpr, quizID sql.NullString
	var enabled bool
	var filtersJSON []byte
	var nextRunAt, lastRunAt sql.NullTime
	var createdAt, updatedAt time.Time

	if err := row.Scan(
		&id, &title, &exportType, &exportFormat, &scheduleType, &cronExpr, &timezone,
		&enabled, &filtersJSON, &quizID, &nextRunAt, &lastRunAt, &createdAt, &updatedAt,
	); err != nil {
		return structs.ScheduledReportResponse{}, err
	}
	return BuildScheduledReportResponse(
		id, title, exportType, exportFormat, scheduleType,
		cronExpr.String, timezone, enabled, json.RawMessage(filtersJSON),
		quizID, nextRunAt, lastRunAt, createdAt, updatedAt,
	), nil
}

func encodeCursorTime(t time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(t.UTC().Format(time.RFC3339Nano)))
}

func decodeCursorTime(cursor string) (time.Time, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, string(b))
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
