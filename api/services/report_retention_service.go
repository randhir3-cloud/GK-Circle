package services

import (
	"context"
	"database/sql"
	"time"

	goqulib "github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
)

const (
	retentionBatchSize   = 100
	retentionTickSeconds = 3600 // 1 hour
)

// ReportRetentionService purges expired reports from storage and clears storage_key.
// It never physically deletes generated_reports rows — audit integrity is preserved.
// It never touches analytics tables.
type ReportRetentionService struct {
	db      *goqulib.Database
	storage StorageProvider
	logger  *zap.Logger
}

// NewReportRetentionService creates a RetentionService.
func NewReportRetentionService(db *goqulib.Database, storage StorageProvider, logger *zap.Logger) *ReportRetentionService {
	return &ReportRetentionService{db: db, storage: storage, logger: logger}
}

// Start launches the retention goroutine. Stops when ctx is cancelled.
func (svc *ReportRetentionService) Start(ctx context.Context) {
	go func() {
		// Run once immediately on startup to clear any backlog.
		svc.tick(ctx)

		ticker := time.NewTicker(retentionTickSeconds * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				svc.tick(ctx)
			}
		}
	}()
}

// tick purges one batch of expired reports.
func (svc *ReportRetentionService) tick(ctx context.Context) {
	rows, err := svc.db.QueryContext(ctx, `
		SELECT id, storage_key, storage_provider, instructor_id
		FROM generated_reports
		WHERE status = 'COMPLETED'
		  AND deleted_at IS NULL
		  AND expires_at < now()
		ORDER BY expires_at
		LIMIT $1
	`, retentionBatchSize)
	if err != nil {
		svc.logger.Error("retention tick query", zap.Error(err))
		return
	}
	defer rows.Close()

	type expiredRow struct {
		id           string
		storageKey   sql.NullString
		provider     sql.NullString
		instructorID string
	}
	var expired []expiredRow
	for rows.Next() {
		var r expiredRow
		if err := rows.Scan(&r.id, &r.storageKey, &r.provider, &r.instructorID); err == nil {
			expired = append(expired, r)
		}
	}

	for _, r := range expired {
		svc.expire(ctx, r.id, r.storageKey, r.instructorID)
	}
}

func (svc *ReportRetentionService) expire(ctx context.Context, reportID string, storageKey sql.NullString, instructorID string) {
	// Delete storage object.
	if storageKey.Valid && storageKey.String != "" {
		if err := svc.storage.Delete(ctx, storageKey.String); err != nil {
			svc.logger.Error("retention: storage delete failed",
				zap.String("report_id", reportID),
				zap.String("key", storageKey.String),
				zap.Error(err),
			)
			// Continue — clear DB key anyway so we don't loop on this row.
		}
	}

	// Clear storage_key; retain row and status.
	_, err := svc.db.ExecContext(ctx,
		`UPDATE generated_reports SET storage_key = NULL WHERE id = $1`,
		reportID,
	)
	if err != nil {
		svc.logger.Error("retention: clear storage_key failed", zap.String("report_id", reportID), zap.Error(err))
		return
	}

	// Audit log — metadata only.
	svc.db.ExecContext(ctx,
		`INSERT INTO export_audit_log (instructor_id, report_id, action, success) VALUES ($1, $2, 'EXPORT_EXPIRED', true)`,
		instructorID, reportID,
	)

	svc.logger.Info("retention: expired report purged", zap.String("report_id", reportID))
}
