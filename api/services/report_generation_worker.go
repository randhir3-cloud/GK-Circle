package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	goqulib "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"go.uber.org/zap"
)

const (
	workerJobChannelBuffer  = 256
	reclaimLookbackSeconds  = 5
	reclaimBatchSize        = 20
)

// ReportWorker manages a goroutine pool that processes export jobs.
// It uses two complementary job-pickup mechanisms:
//  1. Push-based: scheduler sends job IDs to the channel (fast path).
//  2. Pull-based: a reclaim goroutine polls the DB every ReclaimInterval for
//     QUEUED jobs whose channel push was missed (crash-safe durability).
type ReportWorker struct {
	db              *goqulib.Database
	exportSvc       *ExportService
	emailSvc        *ReportEmailService
	jobQueue        chan uuid.UUID
	poolSize        int
	reclaimInterval time.Duration
	logger          *zap.Logger
}

// NewReportWorker creates a worker pool.
func NewReportWorker(
	db *goqulib.Database,
	exportSvc *ExportService,
	emailSvc *ReportEmailService,
	poolSize int,
	reclaimIntervalSeconds int,
	logger *zap.Logger,
) (*ReportWorker, chan uuid.UUID) {
	if poolSize <= 0 {
		poolSize = 3
	}
	if reclaimIntervalSeconds <= 0 {
		reclaimIntervalSeconds = 30
	}
	ch := make(chan uuid.UUID, workerJobChannelBuffer)
	return &ReportWorker{
		db:              db,
		exportSvc:       exportSvc,
		emailSvc:        emailSvc,
		jobQueue:        ch,
		poolSize:        poolSize,
		reclaimInterval: time.Duration(reclaimIntervalSeconds) * time.Second,
		logger:          logger,
	}, ch
}

// Start launches the worker pool and reclaim loop. Stops when ctx is cancelled.
func (w *ReportWorker) Start(ctx context.Context) {
	// Startup recovery — reclaim any jobs from before the current process started.
	w.reclaimQueued(ctx)

	var wg sync.WaitGroup

	// Worker goroutines.
	for i := 0; i < w.poolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case jobID, ok := <-w.jobQueue:
					if !ok {
						return
					}
					w.process(ctx, jobID)
				}
			}
		}()
	}

	// Reclaim loop — polls every ReclaimInterval regardless of channel activity.
	go func() {
		ticker := time.NewTicker(w.reclaimInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.reclaimQueued(ctx)
			}
		}
	}()
}

// reclaimQueued claims QUEUED jobs from the DB that are older than reclaimLookbackSeconds.
// FOR UPDATE SKIP LOCKED ensures only one instance claims each job.
func (w *ReportWorker) reclaimQueued(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM generated_reports
		WHERE status = 'QUEUED'
		  AND queued_at < now() - $1::interval
		ORDER BY queued_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, reclaimLookbackSeconds, reclaimBatchSize)
	if err != nil {
		return
	}

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()

	for _, id := range ids {
		// Claim atomically.
		res, err := tx.ExecContext(ctx,
			`UPDATE generated_reports SET status = 'RUNNING', started_at = now() WHERE id = $1 AND status = 'QUEUED'`,
			id,
		)
		if err != nil {
			continue
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			continue // Another instance claimed it.
		}
		select {
		case w.jobQueue <- id:
		default:
		}
	}
	_ = tx.Commit()
}

// process executes one export job.
func (w *ReportWorker) process(ctx context.Context, jobID uuid.UUID) {
	// Load job metadata.
	row := w.db.QueryRowContext(ctx, `
		SELECT instructor_id, export_type, export_format, filters_json, quiz_id, scheduled_report_id
		FROM generated_reports
		WHERE id = $1
	`, jobID)

	var instructorID, exportType, exportFormat string
	var filtersJSON []byte
	var quizIDStr sql.NullString
	var scheduledReportID sql.NullString

	if err := row.Scan(&instructorID, &exportType, &exportFormat, &filtersJSON, &quizIDStr, &scheduledReportID); err != nil {
		w.logger.Error("worker: load job metadata failed", zap.String("job_id", jobID.String()), zap.Error(err))
		return
	}

	var quizID *uuid.UUID
	if quizIDStr.Valid {
		id, err := uuid.Parse(quizIDStr.String)
		if err == nil {
			quizID = &id
		}
	}

	start := time.Now()
	err := w.exportSvc.Generate(
		ctx,
		jobID,
		instructorID,
		structs.ExportType(exportType),
		structs.ExportFormat(exportFormat),
		json.RawMessage(filtersJSON),
		quizID,
	)

	if err != nil {
		w.logger.Error("worker: export failed",
			zap.String("job_id", jobID.String()),
			zap.Duration("elapsed", time.Since(start)),
			zap.Error(err),
		)
		return
	}

	w.logger.Info("worker: export completed",
		zap.String("job_id", jobID.String()),
		zap.Duration("elapsed", time.Since(start)),
	)

	// Deliver via email if this was a scheduled report.
	if scheduledReportID.Valid && w.emailSvc != nil {
		go w.emailSvc.SendScheduledReport(context.Background(), jobID, instructorID)
	}
}
