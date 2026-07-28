package structs

import "time"

// ExportType enumerates valid export data types.
type ExportType string

const (
	ExportTypePortfolioOverview  ExportType = "PORTFOLIO_OVERVIEW"
	ExportTypeQuizList           ExportType = "QUIZ_LIST"
	ExportTypeLearnerPerformance ExportType = "LEARNER_PERFORMANCE"
	ExportTypeReleaseMonitoring  ExportType = "RELEASE_MONITORING"
	ExportTypeTimeline           ExportType = "TIMELINE"
	ExportTypeQuizSummary        ExportType = "QUIZ_SUMMARY"
	ExportTypeQuizAttempts       ExportType = "QUIZ_ATTEMPTS"
	ExportTypeQuestionMetrics    ExportType = "QUESTION_METRICS"
	ExportTypeEngagement         ExportType = "ENGAGEMENT"
	ExportTypeFullDashboard      ExportType = "FULL_DASHBOARD"
)

// ExportFormat enumerates supported output formats.
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "CSV"
	ExportFormatXLSX ExportFormat = "XLSX"
	ExportFormatPDF  ExportFormat = "PDF"
)

// ScheduleType enumerates scheduling cadences.
type ScheduleType string

const (
	ScheduleTypeDaily    ScheduleType = "DAILY"
	ScheduleTypeWeekly   ScheduleType = "WEEKLY"
	ScheduleTypeMonthly  ScheduleType = "MONTHLY"
	ScheduleTypeOneTime  ScheduleType = "ONE_TIME"
)

// ReportStatus enumerates job lifecycle states.
type ReportStatus string

const (
	ReportStatusQueued    ReportStatus = "QUEUED"
	ReportStatusRunning   ReportStatus = "RUNNING"
	ReportStatusCompleted ReportStatus = "COMPLETED"
	ReportStatusFailed    ReportStatus = "FAILED"
	ReportStatusCancelled ReportStatus = "CANCELLED"
)

// ExportJobResponse is the full job status DTO returned by the status-poll endpoint.
type ExportJobResponse struct {
	ID                 string       `json:"id"`
	Status             ReportStatus `json:"status"`
	ExportType         ExportType   `json:"export_type"`
	ExportFormat       ExportFormat `json:"export_format"`
	Title              string       `json:"title"`
	QuizID             *string      `json:"quiz_id,omitempty"`
	RowCount           *int         `json:"row_count,omitempty"`
	FileSizeBytes      *int64       `json:"file_size_bytes,omitempty"`
	ErrorMessage       *string      `json:"error_message,omitempty"`
	SnapshotStartedAt  *time.Time   `json:"snapshot_started_at,omitempty"`
	QueuedAt           time.Time    `json:"queued_at"`
	StartedAt          *time.Time   `json:"started_at,omitempty"`
	CompletedAt        *time.Time   `json:"completed_at,omitempty"`
	ExpiresAt          *time.Time   `json:"expires_at,omitempty"`
	DeletedAt          *time.Time   `json:"deleted_at,omitempty"`
	ScheduledReportID  *string      `json:"scheduled_report_id,omitempty"`
}

// GeneratedReportListItem is the summary DTO returned by the history list endpoint.
type GeneratedReportListItem struct {
	ID            string       `json:"id"`
	Status        ReportStatus `json:"status"`
	ExportType    ExportType   `json:"export_type"`
	ExportFormat  ExportFormat `json:"export_format"`
	Title         string       `json:"title"`
	QuizID        *string      `json:"quiz_id,omitempty"`
	RowCount      *int         `json:"row_count,omitempty"`
	FileSizeBytes *int64       `json:"file_size_bytes,omitempty"`
	QueuedAt      time.Time    `json:"queued_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	DeletedAt     *time.Time   `json:"deleted_at,omitempty"`
}

// GeneratedReportListResponse wraps a paginated list of history items.
type GeneratedReportListResponse struct {
	Items      []GeneratedReportListItem `json:"items"`
	NextCursor *string                   `json:"next_cursor,omitempty"`
	Total      int                       `json:"total"`
}

// ScheduledReportResponse is the DTO for a scheduled report definition.
type ScheduledReportResponse struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	ExportType   ExportType   `json:"export_type"`
	ExportFormat ExportFormat `json:"export_format"`
	ScheduleType ScheduleType `json:"schedule_type"`
	CronExpr     *string      `json:"cron_expr,omitempty"`
	Timezone     string       `json:"timezone"`
	Enabled      bool         `json:"enabled"`
	FiltersJSON  interface{}  `json:"filters"`
	QuizID       *string      `json:"quiz_id,omitempty"`
	NextRunAt    *time.Time   `json:"next_run_at,omitempty"`
	LastRunAt    *time.Time   `json:"last_run_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// ScheduledReportListResponse wraps a list of schedule definitions.
type ScheduledReportListResponse struct {
	Items []ScheduledReportResponse `json:"items"`
}

// AuditLogItem is the DTO for a single audit log entry.
type AuditLogItem struct {
	ID            string     `json:"id"`
	Action        string     `json:"action"`
	ExportType    *string    `json:"export_type,omitempty"`
	ExportFormat  *string    `json:"export_format,omitempty"`
	ReportID      *string    `json:"report_id,omitempty"`
	FiltersJSON   interface{} `json:"filters,omitempty"`
	DurationMs    *int       `json:"duration_ms,omitempty"`
	RowCount      *int       `json:"row_count,omitempty"`
	Success       *bool      `json:"success,omitempty"`
	OccurredAt    time.Time  `json:"occurred_at"`
	CorrelationID *string    `json:"correlation_id,omitempty"`
}

// AuditLogListResponse wraps a paginated list of audit entries.
type AuditLogListResponse struct {
	Items      []AuditLogItem `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

// RequestExportInput is the body for POST /instructor/reports/exports.
type RequestExportInput struct {
	ExportType   ExportType   `json:"export_type"`
	ExportFormat ExportFormat `json:"export_format"`
	Title        string       `json:"title"`
	FiltersJSON  interface{}  `json:"filters"`
	QuizID       *string      `json:"quiz_id,omitempty"`
}

// CreateScheduleInput is the body for POST /instructor/reports/schedules.
type CreateScheduleInput struct {
	Title        string       `json:"title"`
	ExportType   ExportType   `json:"export_type"`
	ExportFormat ExportFormat `json:"export_format"`
	ScheduleType ScheduleType `json:"schedule_type"`
	CronExpr     *string      `json:"cron_expr,omitempty"` // required unless ONE_TIME
	Timezone     string       `json:"timezone"`
	FiltersJSON  interface{}  `json:"filters"`
	QuizID       *string      `json:"quiz_id,omitempty"`
}

// UpdateScheduleInput is the body for PATCH /instructor/reports/schedules/:id.
type UpdateScheduleInput struct {
	Title        *string       `json:"title,omitempty"`
	CronExpr     *string       `json:"cron_expr,omitempty"`
	Timezone     *string       `json:"timezone,omitempty"`
	Enabled      *bool         `json:"enabled,omitempty"`
	FiltersJSON  interface{}   `json:"filters,omitempty"`
}
