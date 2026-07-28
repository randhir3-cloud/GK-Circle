package services

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	goqulib "github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

const (
	// retryBackoffDurations defines wait durations for SMTP temporary failure retries.
	emailRetryAttempts = 3
)

var emailRetryBackoff = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	30 * time.Minute,
}

// ReportEmailService sends generated report files to instructors via email.
// It wraps the existing EmailService and extends it with retry logic and audit logging.
// Audit records contain only metadata — never URLs, email body, or attachment paths.
type ReportEmailService struct {
	emailSvc           *EmailService
	db                 *goqulib.Database
	storage            StorageProvider
	maxAttachmentBytes int64
	logger             *zap.Logger
}

// NewReportEmailService creates a ReportEmailService.
func NewReportEmailService(
	emailSvc *EmailService,
	db *goqulib.Database,
	storage StorageProvider,
	maxAttachmentBytes int64,
	logger *zap.Logger,
) *ReportEmailService {
	if maxAttachmentBytes <= 0 {
		maxAttachmentBytes = 10 * 1024 * 1024 // 10 MB default
	}
	return &ReportEmailService{
		emailSvc:           emailSvc,
		db:                 db,
		storage:            storage,
		maxAttachmentBytes: maxAttachmentBytes,
		logger:             logger,
	}
}

// SendScheduledReport delivers a completed scheduled report to its instructor.
// It reads the instructor's email from Kratos identity data, chooses attach vs. link,
// and applies the retry policy.
func (svc *ReportEmailService) SendScheduledReport(ctx context.Context, reportID uuid.UUID, instructorID string) {
	// Load report metadata.
	row := svc.db.QueryRowContext(ctx, `
		SELECT title, export_format, storage_key, file_size_bytes, expires_at
		FROM generated_reports
		WHERE id = $1 AND status = 'COMPLETED' AND deleted_at IS NULL
	`, reportID)

	var title, exportFormat string
	var storageKey sql.NullString
	var fileSizeBytes sql.NullInt64
	var expiresAt sql.NullTime

	if err := row.Scan(&title, &exportFormat, &storageKey, &fileSizeBytes, &expiresAt); err != nil {
		svc.logger.Error("report email: load report failed", zap.String("report_id", reportID.String()), zap.Error(err))
		return
	}

	// Check expiry before attempting delivery.
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		svc.logger.Warn("report email: report already expired, skipping", zap.String("report_id", reportID.String()))
		svc.logDelivery(ctx, reportID, instructorID, "FAILED", 1, nil, "report already expired before delivery")
		svc.writeEmailAudit(ctx, instructorID, &reportID, "EMAIL_FAILED", false)
		return
	}

	// Retrieve instructor email — read from the users table (populated by Kratos sync).
	var recipientEmail string
	emailRow := svc.db.QueryRowContext(ctx,
		`SELECT COALESCE(email, '') FROM users WHERE id = $1`, instructorID)
	if err := emailRow.Scan(&recipientEmail); err != nil || recipientEmail == "" {
		svc.logger.Error("report email: could not resolve instructor email",
			zap.String("instructor_id", instructorID), zap.Error(err))
		return
	}

	svc.sendWithRetry(ctx, reportID, instructorID, recipientEmail, title, exportFormat, storageKey, fileSizeBytes, expiresAt)
}

// sendWithRetry attempts delivery up to emailRetryAttempts times for temporary failures.
func (svc *ReportEmailService) sendWithRetry(
	ctx context.Context,
	reportID uuid.UUID,
	instructorID, recipientEmail, title, exportFormat string,
	storageKey sql.NullString,
	fileSizeBytes sql.NullInt64,
	expiresAt sql.NullTime,
) {
	var lastErr error
	for attempt := 1; attempt <= emailRetryAttempts; attempt++ {
		err := svc.trySend(ctx, reportID, recipientEmail, title, exportFormat, storageKey, fileSizeBytes, expiresAt)
		if err == nil {
			svc.logDelivery(ctx, reportID, instructorID, "SENT", attempt, nil, "")
			svc.writeEmailAudit(ctx, instructorID, &reportID, "EMAIL_SENT", true)
			return
		}

		lastErr = err
		isPermanent := isPermanentSMTPError(err)
		if isPermanent {
			svc.logger.Error("report email: permanent SMTP failure, no retry",
				zap.String("report_id", reportID.String()),
				zap.String("error_class", smtpErrorClass(err)),
			)
			break
		}

		if attempt < emailRetryAttempts {
			backoff := emailRetryBackoff[attempt-1]
			nextRetry := time.Now().Add(backoff)
			svc.logDelivery(ctx, reportID, instructorID, "RETRYING", attempt, &nextRetry, smtpErrorClass(err))
			svc.writeEmailAudit(ctx, instructorID, &reportID, "EMAIL_RETRIED", false)
			svc.logger.Warn("report email: temporary failure, will retry",
				zap.String("report_id", reportID.String()),
				zap.Duration("backoff", backoff),
				zap.Int("attempt", attempt),
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
		}
	}

	// All attempts exhausted.
	svc.logger.Error("report email: all attempts failed",
		zap.String("report_id", reportID.String()),
		zap.Error(lastErr),
	)
	svc.logDelivery(ctx, reportID, instructorID, "FAILED", emailRetryAttempts, nil, smtpErrorClass(lastErr))
	svc.writeEmailAudit(ctx, instructorID, &reportID, "EMAIL_FAILED", false)
}

// trySend performs a single delivery attempt.
func (svc *ReportEmailService) trySend(
	ctx context.Context,
	reportID uuid.UUID,
	recipientEmail, title, exportFormat string,
	storageKey sql.NullString,
	fileSizeBytes sql.NullInt64,
	expiresAt sql.NullTime,
) error {
	// Decide: attach directly or send download link.
	if fileSizeBytes.Valid && fileSizeBytes.Int64 <= svc.maxAttachmentBytes && storageKey.Valid {
		return svc.sendWithAttachment(ctx, recipientEmail, title, exportFormat, storageKey.String, expiresAt)
	}
	return svc.sendWithLink(ctx, reportID, recipientEmail, title, expiresAt)
}

func (svc *ReportEmailService) sendWithAttachment(
	ctx context.Context,
	recipientEmail, title, exportFormat, key string,
	expiresAt sql.NullTime,
) error {
	rc, _, err := svc.storage.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("storage get for attachment: %w", err)
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return fmt.Errorf("read attachment: %w", err)
	}

	ext := strings.ToLower(exportFormat)
	filename := fmt.Sprintf("report.%s", ext)
	contentType := ContentTypeForFormat(exportFormat)

	expiry := ""
	if expiresAt.Valid {
		expiry = fmt.Sprintf("<p>This report is available until %s.</p>", expiresAt.Time.UTC().Format("2006-01-02 15:04 UTC"))
	}
	body := fmt.Sprintf(`<h2>%s</h2><p>Your report is attached.</p>%s`, title, expiry)

	m := gomail.NewMessage()
	m.SetHeader("From", svc.emailSvc.config.EmailFrom)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", fmt.Sprintf("GK Circle Report: %s", title))
	m.SetBody("text/html", body)
	m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(buf.Bytes())
		return err
	}), gomail.SetHeader(map[string][]string{"Content-Type": {contentType}}))

	return svc.emailSvc.dialer.DialAndSend(m)
}

func (svc *ReportEmailService) sendWithLink(
	ctx context.Context,
	reportID uuid.UUID,
	recipientEmail, title string,
	expiresAt sql.NullTime,
) error {
	// We send an opaque download URL — the actual signed URL is NOT logged.
	downloadPath := fmt.Sprintf("/instructor/reports/%s", reportID.String())
	expiry := ""
	if expiresAt.Valid {
		expiry = fmt.Sprintf("<p>Link valid until: %s UTC.</p>", expiresAt.Time.UTC().Format("2006-01-02 15:04"))
	}
	body := fmt.Sprintf(
		`<h2>%s</h2><p>Your report is ready. <a href="%s">Download here</a>.</p>%s`,
		title, downloadPath, expiry,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", svc.emailSvc.config.EmailFrom)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", fmt.Sprintf("GK Circle Report Ready: %s", title))
	m.SetBody("text/html", body)

	return svc.emailSvc.dialer.DialAndSend(m)
}

// logDelivery writes to report_delivery_logs. Stores only error class — no body, no URLs.
func (svc *ReportEmailService) logDelivery(
	ctx context.Context,
	reportID uuid.UUID,
	instructorID, status string,
	attempt int,
	nextRetry *time.Time,
	errClass string,
) {
	svc.db.ExecContext(ctx, `
		INSERT INTO report_delivery_logs
			(report_id, recipient_email, delivery_status, attempt_number, next_retry_at, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, reportID, instructorID, status, attempt, nextRetry, nullStr(errClass))
}

// writeEmailAudit writes to export_audit_log — metadata only.
func (svc *ReportEmailService) writeEmailAudit(ctx context.Context, instructorID string, reportID *uuid.UUID, action string, success bool) {
	var rID *string
	if reportID != nil {
		s := reportID.String()
		rID = &s
	}
	svc.db.ExecContext(ctx,
		`INSERT INTO export_audit_log (instructor_id, report_id, action, success) VALUES ($1, $2, $3, $4)`,
		instructorID, rID, action, success,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// SMTP error classification helpers
// ─────────────────────────────────────────────────────────────────────────────

// isPermanentSMTPError returns true for 5xx SMTP errors (permanent failures).
func isPermanentSMTPError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// gomail surfaces SMTP response codes in error strings.
	for _, prefix := range []string{"550", "551", "552", "553", "554", "500", "501", "502", "503"} {
		if strings.Contains(msg, prefix) {
			return true
		}
	}
	return false
}

// smtpErrorClass extracts a safe error class for logging (no personal data, no URLs).
func smtpErrorClass(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "connection refused"):
		return "CONNECTION_REFUSED"
	case strings.Contains(msg, "timeout"):
		return "TIMEOUT"
	case strings.Contains(msg, "TLS"):
		return "TLS_ERROR"
	case strings.Contains(msg, "550"):
		return "SMTP_5XX"
	case strings.Contains(msg, "4"):
		return "SMTP_4XX"
	default:
		return "SMTP_ERROR"
	}
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
