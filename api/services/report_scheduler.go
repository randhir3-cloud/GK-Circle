package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	goqulib "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const schedulerMaxBatch = 50

// ReportScheduler polls the database for due scheduled reports and enqueues them.
// It uses FOR UPDATE SKIP LOCKED to be safe across multiple API instances.
type ReportScheduler struct {
	db       *goqulib.Database
	jobQueue chan<- uuid.UUID
	interval time.Duration
	logger   *zap.Logger
}

// NewReportScheduler creates a scheduler that sends due job IDs to jobQueue.
func NewReportScheduler(db *goqulib.Database, jobQueue chan<- uuid.UUID, intervalSeconds int, logger *zap.Logger) *ReportScheduler {
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}
	return &ReportScheduler{
		db:       db,
		jobQueue: jobQueue,
		interval: time.Duration(intervalSeconds) * time.Second,
		logger:   logger,
	}
}

// Start launches the scheduler goroutine. It stops when ctx is cancelled.
func (s *ReportScheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tick(ctx)
			}
		}
	}()
}

// tick claims and dispatches all due schedules atomically.
func (s *ReportScheduler) tick(ctx context.Context) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		s.logger.Error("scheduler tx begin", zap.Error(err))
		return
	}
	defer tx.Rollback()

	// FOR UPDATE SKIP LOCKED prevents duplicate dispatch across concurrent instances.
	rows, err := tx.QueryContext(ctx, `
		SELECT id, instructor_id, export_type, export_format, schedule_type,
		       cron_expr, timezone, filters_json, quiz_id
		FROM scheduled_reports
		WHERE enabled = true AND next_run_at <= now()
		ORDER BY next_run_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, schedulerMaxBatch)
	if err != nil {
		s.logger.Error("scheduler query", zap.Error(err))
		return
	}

	type dueRow struct {
		id           string
		instructorID string
		exportType   string
		exportFormat string
		scheduleType string
		cronExpr     sql.NullString
		timezone     string
		filtersJSON  []byte
		quizID       sql.NullString
	}
	var due []dueRow
	for rows.Next() {
		var r dueRow
		if err := rows.Scan(&r.id, &r.instructorID, &r.exportType, &r.exportFormat,
			&r.scheduleType, &r.cronExpr, &r.timezone, &r.filtersJSON, &r.quizID); err != nil {
			s.logger.Error("scheduler row scan", zap.Error(err))
		} else {
			due = append(due, r)
		}
	}
	rows.Close()

	for _, r := range due {
		jobID := uuid.New()
		title := fmt.Sprintf("%s %s (scheduled)", r.exportType, r.exportFormat)

		var quizIDArg interface{} = nil
		if r.quizID.Valid {
			quizIDArg = r.quizID.String
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO generated_reports
				(id, scheduled_report_id, instructor_id, title, export_type, export_format, status, filters_json, quiz_id)
			VALUES ($1, $2, $3, $4, $5, $6, 'QUEUED', $7, $8)
		`, jobID, r.id, r.instructorID, title, r.exportType, r.exportFormat, r.filtersJSON, quizIDArg)
		if err != nil {
			s.logger.Error("scheduler insert job", zap.Error(err), zap.String("schedule_id", r.id))
			continue
		}

		// Compute next run.
		nextRun, enabled := computeNextRun(r.scheduleType, r.cronExpr.String, r.timezone)

		_, err = tx.ExecContext(ctx,
			`UPDATE scheduled_reports SET last_run_at = now(), next_run_at = $2, enabled = $3 WHERE id = $1`,
			r.id, nextRun, enabled,
		)
		if err != nil {
			s.logger.Error("scheduler update schedule", zap.Error(err))
		}

		s.logger.Info("scheduler dispatched job",
			zap.String("schedule_id", r.id),
			zap.String("job_id", jobID.String()),
		)

		// Push to channel post-commit (best-effort fast path).
		// If lost, the reclaim loop picks it up within ReclaimIntervalSeconds.
		select {
		case s.jobQueue <- jobID:
		default:
			// Channel full — reclaim loop will handle it.
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("scheduler tx commit", zap.Error(err))
	}
}

// computeNextRun returns the next scheduled time and whether the schedule remains enabled.
func computeNextRun(scheduleType, cronExpr, timezone string) (*time.Time, bool) {
	if scheduleType == string(structs.ScheduleTypeOneTime) {
		return nil, false // Disable after first run.
	}
	if cronExpr == "" {
		return nil, false
	}
	loc := resolveTimezone(timezone)
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, false
	}
	t := schedule.Next(time.Now().In(loc))
	return &t, true
}

// resolveTimezone applies T02/T03 timezone resolution: IANA → platform default → UTC.
func resolveTimezone(tz string) *time.Location {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	// Platform default (could be read from config; for now fall back to UTC).
	return time.UTC
}

// ValidateCronExpr validates a cron expression at request time.
// Returns an error suitable for a 400 response if the expression is invalid.
func ValidateCronExpr(expr string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	return err
}

// ValidateIANATimezone validates a timezone string against Go's embedded IANA database.
func ValidateIANATimezone(tz string) error {
	if tz == "" {
		return fmt.Errorf("timezone is required")
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("invalid IANA timezone %q: %w", tz, err)
	}
	return nil
}

// BuildScheduledReportResponse converts a DB row to the DTO.
func BuildScheduledReportResponse(
	id, title, exportType, exportFormat, scheduleType, cronExpr, timezone string,
	enabled bool,
	filtersJSON json.RawMessage,
	quizID sql.NullString,
	nextRunAt, lastRunAt sql.NullTime,
	createdAt, updatedAt time.Time,
) structs.ScheduledReportResponse {
	resp := structs.ScheduledReportResponse{
		ID:           id,
		Title:        title,
		ExportType:   structs.ExportType(exportType),
		ExportFormat: structs.ExportFormat(exportFormat),
		ScheduleType: structs.ScheduleType(scheduleType),
		Timezone:     timezone,
		Enabled:      enabled,
		FiltersJSON:  json.RawMessage(filtersJSON),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	if cronExpr != "" {
		resp.CronExpr = &cronExpr
	}
	if quizID.Valid {
		resp.QuizID = &quizID.String
	}
	if nextRunAt.Valid {
		t := nextRunAt.Time
		resp.NextRunAt = &t
	}
	if lastRunAt.Valid {
		t := lastRunAt.Time
		resp.LastRunAt = &t
	}
	return resp
}
