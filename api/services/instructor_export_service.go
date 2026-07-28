package services

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	goqulib "github.com/doug-martin/goqu/v9"
	"github.com/phpdave11/gofpdf"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

const (
	exportBatchSize    = 200
	exportRetentionTTL = 30 * 24 * time.Hour // overridden by config.ReportConfig.RetentionDays
	pdfPageMarginMM    = 15.0
	pdfFontSize        = 9.0
	pdfHeaderFontSize  = 11.0
)

// ExportService generates CSV, XLSX, and PDF report files from T03 analytics data.
// It is strictly read-only: it never modifies analytics tables or bumps analytics caches.
type ExportService struct {
	db              *goqulib.Database
	analyticsSvc    *InstructorAnalyticsService
	storage         StorageProvider
	retentionDays   int
	logger          *zap.Logger
}

// NewExportService creates an ExportService.
func NewExportService(
	db *goqulib.Database,
	analyticsSvc *InstructorAnalyticsService,
	storage StorageProvider,
	retentionDays int,
	logger *zap.Logger,
) *ExportService {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return &ExportService{
		db:            db,
		analyticsSvc:  analyticsSvc,
		storage:       storage,
		retentionDays: retentionDays,
		logger:        logger,
	}
}

// Generate dispatches to the appropriate format generator.
func (svc *ExportService) Generate(ctx context.Context, reportID uuid.UUID, instructorID string, exportType structs.ExportType, exportFormat structs.ExportFormat, filtersJSON json.RawMessage, quizID *uuid.UUID) error {
	switch exportFormat {
	case structs.ExportFormatCSV:
		return svc.GenerateCSV(ctx, reportID, instructorID, exportType, filtersJSON, quizID)
	case structs.ExportFormatXLSX:
		return svc.GenerateXLSX(ctx, reportID, instructorID, exportType, filtersJSON, quizID)
	case structs.ExportFormatPDF:
		return svc.GeneratePDF(ctx, reportID, instructorID, exportType, filtersJSON, quizID)
	default:
		return svc.markFailed(ctx, reportID, fmt.Sprintf("unsupported export format: %s", exportFormat))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CSV
// ─────────────────────────────────────────────────────────────────────────────

func (svc *ExportService) GenerateCSV(ctx context.Context, reportID uuid.UUID, instructorID string, exportType structs.ExportType, filtersJSON json.RawMessage, quizID *uuid.UUID) error {
	if err := svc.markRunning(ctx, reportID); err != nil {
		return err
	}

	snapshotAt := time.Now().UTC()
	pr, pw := io.Pipe()
	var rowCount int
	var writeErr error

	go func() {
		defer pw.Close()
		cw := csv.NewWriter(pw)
		defer cw.Flush()

		// Open REPEATABLE READ transaction for snapshot consistency.
		tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		defer tx.Rollback()

		headers, rows, err := svc.streamRows(ctx, tx, exportType, instructorID, filtersJSON, quizID)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := cw.Write(headers); err != nil {
			pw.CloseWithError(err)
			return
		}
		for row := range rows {
			if err := cw.Write(row); err != nil {
				writeErr = err
				return
			}
			rowCount++
		}
		if writeErr != nil {
			pw.CloseWithError(writeErr)
			return
		}
		tx.Commit()
	}()

	storageKey := storageKey(reportID, "csv")
	// Size unknown during streaming; use -1 to indicate chunked.
	if err := svc.storage.Put(ctx, storageKey, pr, -1, ContentTypeForFormat("CSV")); err != nil {
		_ = svc.markFailed(ctx, reportID, err.Error())
		return err
	}

	return svc.markCompleted(ctx, reportID, storageKey, "local", rowCount, snapshotAt)
}

// ─────────────────────────────────────────────────────────────────────────────
// XLSX
// ─────────────────────────────────────────────────────────────────────────────

func (svc *ExportService) GenerateXLSX(ctx context.Context, reportID uuid.UUID, instructorID string, exportType structs.ExportType, filtersJSON json.RawMessage, quizID *uuid.UUID) error {
	if err := svc.markRunning(ctx, reportID); err != nil {
		return err
	}

	snapshotAt := time.Now().UTC()
	pr, pw := io.Pipe()
	var rowCount int

	go func() {
		defer pw.Close()

		f := excelize.NewFile()
		defer f.Close()

		sheet := "Report"
		f.SetSheetName("Sheet1", sheet)
		sw, err := f.NewStreamWriter(sheet)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		defer tx.Rollback()

		headers, rows, err := svc.streamRows(ctx, tx, exportType, instructorID, filtersJSON, quizID)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		// Write header row with bold style.
		styleID, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		headerCells := make([]interface{}, len(headers))
		for i, h := range headers {
			headerCells[i] = excelize.Cell{StyleID: styleID, Value: h}
		}
		_ = sw.SetRow("A1", headerCells, excelize.RowOpts{Height: 16})

		// Freeze header row.
		_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

		rowNum := 2
		for row := range rows {
			cells := make([]interface{}, len(row))
			for i, v := range row {
				cells[i] = v
			}
			cell, _ := excelize.CoordinatesToCellName(1, rowNum)
			_ = sw.SetRow(cell, cells)
			rowNum++
			rowCount++
		}
		if err := sw.Flush(); err != nil {
			pw.CloseWithError(err)
			return
		}
		tx.Commit()
		if _, err := f.WriteTo(pw); err != nil {
			pw.CloseWithError(err)
		}
	}()

	storageKey := storageKey(reportID, "xlsx")
	if err := svc.storage.Put(ctx, storageKey, pr, -1, ContentTypeForFormat("XLSX")); err != nil {
		_ = svc.markFailed(ctx, reportID, err.Error())
		return err
	}
	return svc.markCompleted(ctx, reportID, storageKey, "local", rowCount, snapshotAt)
}

// ─────────────────────────────────────────────────────────────────────────────
// PDF
// ─────────────────────────────────────────────────────────────────────────────

func (svc *ExportService) GeneratePDF(ctx context.Context, reportID uuid.UUID, instructorID string, exportType structs.ExportType, filtersJSON json.RawMessage, quizID *uuid.UUID) error {
	if err := svc.markRunning(ctx, reportID); err != nil {
		return err
	}

	snapshotAt := time.Now().UTC()

	tx, err := svc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		_ = svc.markFailed(ctx, reportID, err.Error())
		return err
	}
	defer tx.Rollback()

	headers, rows, err := svc.streamRows(ctx, tx, exportType, instructorID, filtersJSON, quizID)
	if err != nil {
		_ = svc.markFailed(ctx, reportID, err.Error())
		return err
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(pdfPageMarginMM, pdfPageMarginMM, pdfPageMarginMM)
	pdf.SetAutoPageBreak(true, pdfPageMarginMM)

	// Header function — called on every new page.
	colWidth := (297.0 - 2*pdfPageMarginMM) / float64(max(len(headers), 1))
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Helvetica", "B", pdfHeaderFontSize)
		pdf.SetFillColor(60, 90, 153)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetY(pdfPageMarginMM)
		for _, h := range headers {
			pdf.CellFormat(colWidth, 8, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(8)
		pdf.SetFillColor(255, 255, 255)
		pdf.SetTextColor(0, 0, 0)
	})

	// Footer function — called on every page.
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d / {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("{nb}")

	pdf.AddPage()

	// Cover information block before table.
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 10, fmt.Sprintf("GK Circle — %s Report", exportTypeLabel(exportType)), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(0, 6, fmt.Sprintf("Generated: %s UTC", snapshotAt.Format("2006-01-02 15:04")), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Instructor: %s", instructorID), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	rowCount := 0
	fill := false
	pdf.SetFont("Helvetica", "", pdfFontSize)
	for row := range rows {
		_, pageHt := pdf.GetPageSize()
		if pdf.GetY()+8 > pageHt-pdfPageMarginMM {
			pdf.AddPage()
		}
		if fill {
			pdf.SetFillColor(245, 245, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		for _, cell := range row {
			pdf.CellFormat(colWidth, 7, truncateStr(cell, 30), "1", 0, "L", true, 0, "")
		}
		pdf.Ln(7)
		fill = !fill
		rowCount++
	}

	tx.Commit()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := pdf.Output(pw); err != nil {
			pw.CloseWithError(err)
		}
	}()

	storageKey := storageKey(reportID, "pdf")
	if err := svc.storage.Put(ctx, storageKey, pr, -1, ContentTypeForFormat("PDF")); err != nil {
		_ = svc.markFailed(ctx, reportID, err.Error())
		return err
	}
	return svc.markCompleted(ctx, reportID, storageKey, "local", rowCount, snapshotAt)
}

// ─────────────────────────────────────────────────────────────────────────────
// Row streaming — delegates to T03 InstructorAnalyticsService
// ─────────────────────────────────────────────────────────────────────────────

// streamRows returns a header row and a channel of string rows for the given export type.
// All queries run inside the provided tx (REPEATABLE READ).
func (svc *ExportService) streamRows(
	ctx context.Context,
	tx *goqulib.TxDatabase,
	exportType structs.ExportType,
	instructorID string,
	filtersJSON json.RawMessage,
	quizID *uuid.UUID,
) ([]string, <-chan []string, error) {
	switch exportType {
	case structs.ExportTypePortfolioOverview:
		return svc.streamPortfolioOverview(ctx, tx, instructorID)
	case structs.ExportTypeQuizList:
		return svc.streamQuizList(ctx, tx, instructorID, filtersJSON)
	case structs.ExportTypeLearnerPerformance:
		return svc.streamLearnerPerformance(ctx, tx, instructorID, filtersJSON)
	case structs.ExportTypeReleaseMonitoring:
		return svc.streamReleaseMonitoring(ctx, tx, instructorID, filtersJSON)
	case structs.ExportTypeQuizSummary:
		if quizID == nil {
			return nil, nil, errors.New("quiz_id required for QUIZ_SUMMARY export")
		}
		return svc.streamQuizSummary(ctx, tx, quizID.String())
	case structs.ExportTypeQuizAttempts:
		if quizID == nil {
			return nil, nil, errors.New("quiz_id required for QUIZ_ATTEMPTS export")
		}
		return svc.streamQuizAttempts(ctx, tx, quizID.String(), filtersJSON)
	case structs.ExportTypeQuestionMetrics:
		if quizID == nil {
			return nil, nil, errors.New("quiz_id required for QUESTION_METRICS export")
		}
		return svc.streamQuestionMetrics(ctx, tx, quizID.String(), filtersJSON)
	default:
		return nil, nil, fmt.Errorf("unsupported export type: %s", exportType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Per-type streaming implementations
// Each opens a paginated cursor within the tx and pushes rows to a channel.
// ─────────────────────────────────────────────────────────────────────────────

func (svc *ExportService) streamPortfolioOverview(ctx context.Context, tx *goqulib.TxDatabase, instructorID string) ([]string, <-chan []string, error) {
	headers := []string{"Total Quizzes", "Published Quizzes", "Total Attempts", "Completed Attempts", "Average Score %", "Active Learners"}
	ch := make(chan []string)
	go func() {
		defer close(ch)
		// Single-row summary from T03 overview query.
		row := tx.QueryRowContext(ctx, `
			SELECT
				COUNT(DISTINCT q.id) FILTER (WHERE q.instructor_id = $1) AS total_quizzes,
				COUNT(DISTINCT q.id) FILTER (WHERE q.instructor_id = $1 AND q.is_public = true) AS published_quizzes,
				COUNT(DISTINCT aa.id) FILTER (WHERE q.instructor_id = $1) AS total_attempts,
				COUNT(DISTINCT aa.id) FILTER (WHERE q.instructor_id = $1 AND aa.status IN ('submitted','finalised')) AS completed_attempts,
				ROUND(AVG(aa.percentage) FILTER (WHERE q.instructor_id = $1 AND aa.status IN ('submitted','finalised')), 2) AS avg_score,
				COUNT(DISTINCT aa.user_id) FILTER (WHERE q.instructor_id = $1 AND aa.started_at >= now() - interval '30 days') AS active_learners
			FROM quizzes q
			LEFT JOIN assessment_attempts aa ON aa.quiz_id = q.id
		`, instructorID)
		var total, published, attempts, completed, active sql.NullInt64
		var avg sql.NullFloat64
		_ = row.Scan(&total, &published, &attempts, &completed, &avg, &active)
		ch <- []string{
			nullInt64Str(total), nullInt64Str(published),
			nullInt64Str(attempts), nullInt64Str(completed),
			nullFloat64Str(avg), nullInt64Str(active),
		}
	}()
	return headers, ch, nil
}

func (svc *ExportService) streamQuizList(ctx context.Context, tx *goqulib.TxDatabase, instructorID string, _ json.RawMessage) ([]string, <-chan []string, error) {
	headers := []string{"Quiz ID", "Title", "Is Public", "Total Attempts", "Completed", "Avg Score %", "Created At"}
	ch := make(chan []string, exportBatchSize)
	go func() {
		defer close(ch)
		rows, err := tx.QueryContext(ctx, `
			SELECT q.id, q.title, q.is_public,
				COUNT(aa.id) AS total_attempts,
				COUNT(aa.id) FILTER (WHERE aa.status IN ('submitted','finalised')) AS completed,
				ROUND(AVG(aa.percentage) FILTER (WHERE aa.status IN ('submitted','finalised')), 2) AS avg_score,
				q.created_at
			FROM quizzes q
			LEFT JOIN assessment_attempts aa ON aa.quiz_id = q.id
			WHERE q.instructor_id = $1
			GROUP BY q.id, q.title, q.is_public, q.created_at
			ORDER BY q.created_at DESC
		`, instructorID)
		if err != nil {
			svc.logger.Error("export streamQuizList query", zap.Error(err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id, title string
			var isPublic bool
			var total, completed sql.NullInt64
			var avg sql.NullFloat64
			var createdAt time.Time
			if err := rows.Scan(&id, &title, &isPublic, &total, &completed, &avg, &createdAt); err != nil {
				return
			}
			ch <- []string{id, title, boolStr(isPublic), nullInt64Str(total), nullInt64Str(completed), nullFloat64Str(avg), createdAt.Format(time.RFC3339)}
		}
	}()
	return headers, ch, nil
}

func (svc *ExportService) streamLearnerPerformance(ctx context.Context, tx *goqulib.TxDatabase, instructorID string, filtersJSON json.RawMessage) ([]string, <-chan []string, error) {
	headers := []string{"User ID", "Display Name", "Unique Quizzes", "Total Attempts", "Completed", "Completion Rate %", "Avg Score %", "Last Activity"}
	ch := make(chan []string, exportBatchSize)

	var quizIDFilter *string
	if len(filtersJSON) > 0 {
		var f struct {
			QuizID *string `json:"quiz_id"`
		}
		_ = json.Unmarshal(filtersJSON, &f)
		quizIDFilter = f.QuizID
	}

	go func() {
		defer close(ch)
		query := `
			SELECT
				aa.user_id,
				COALESCE(u.first_name || ' ' || u.last_name, aa.user_id) AS display_name,
				COUNT(DISTINCT aa.quiz_id) AS unique_quizzes,
				COUNT(aa.id) AS total_attempts,
				COUNT(aa.id) FILTER (WHERE aa.status IN ('submitted','finalised')) AS completed,
				ROUND(
					100.0 * COUNT(aa.id) FILTER (WHERE aa.status IN ('submitted','finalised'))
					/ NULLIF(COUNT(aa.id), 0), 2
				) AS completion_rate,
				ROUND(AVG(aa.percentage) FILTER (WHERE aa.status IN ('submitted','finalised')), 2) AS avg_score,
				MAX(aa.started_at) AS last_activity
			FROM assessment_attempts aa
			JOIN quizzes q ON q.id = aa.quiz_id
			LEFT JOIN users u ON u.id = aa.user_id
			WHERE q.instructor_id = $1
		`
		args := []interface{}{instructorID}
		if quizIDFilter != nil {
			query += ` AND aa.quiz_id = $2`
			args = append(args, *quizIDFilter)
		}
		query += ` GROUP BY aa.user_id, display_name ORDER BY last_activity DESC`

		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			svc.logger.Error("export streamLearnerPerformance query", zap.Error(err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var userID, displayName string
			var unique, total, completed sql.NullInt64
			var compRate, avgScore sql.NullFloat64
			var lastActivity sql.NullTime
			if err := rows.Scan(&userID, &displayName, &unique, &total, &completed, &compRate, &avgScore, &lastActivity); err != nil {
				return
			}
			ch <- []string{
				userID, displayName,
				nullInt64Str(unique), nullInt64Str(total), nullInt64Str(completed),
				nullFloat64Str(compRate), nullFloat64Str(avgScore),
				nullTimeStr(lastActivity),
			}
		}
	}()
	return headers, ch, nil
}

func (svc *ExportService) streamReleaseMonitoring(ctx context.Context, tx *goqulib.TxDatabase, instructorID string, _ json.RawMessage) ([]string, <-chan []string, error) {
	headers := []string{"Quiz ID", "Quiz Title", "Release Status", "Scheduled At", "Released At", "Total Attempts", "Results Visible"}
	ch := make(chan []string, exportBatchSize)
	go func() {
		defer close(ch)
		rows, err := tx.QueryContext(ctx, `
			SELECT q.id, q.title,
				COALESCE(qrs.release_status, 'UNRELEASED') AS release_status,
				qrs.scheduled_release_at,
				qrs.released_at,
				COUNT(aa.id) AS total_attempts,
				COALESCE(qrs.results_visible, false) AS results_visible
			FROM quizzes q
			LEFT JOIN quiz_result_settings qrs ON qrs.quiz_id = q.id
			LEFT JOIN assessment_attempts aa ON aa.quiz_id = q.id
			WHERE q.instructor_id = $1
			GROUP BY q.id, q.title, qrs.release_status, qrs.scheduled_release_at, qrs.released_at, qrs.results_visible
			ORDER BY q.created_at DESC
		`, instructorID)
		if err != nil {
			svc.logger.Error("export streamReleaseMonitoring query", zap.Error(err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id, title, status string
			var scheduledAt, releasedAt sql.NullTime
			var total sql.NullInt64
			var visible bool
			if err := rows.Scan(&id, &title, &status, &scheduledAt, &releasedAt, &total, &visible); err != nil {
				return
			}
			ch <- []string{id, title, status, nullTimeStr(scheduledAt), nullTimeStr(releasedAt), nullInt64Str(total), boolStr(visible)}
		}
	}()
	return headers, ch, nil
}

func (svc *ExportService) streamQuizSummary(ctx context.Context, tx *goqulib.TxDatabase, quizID string) ([]string, <-chan []string, error) {
	headers := []string{"Metric", "Value"}
	ch := make(chan []string)
	go func() {
		defer close(ch)
		row := tx.QueryRowContext(ctx, `
			SELECT
				q.title, q.is_public,
				COUNT(aa.id) AS total_attempts,
				COUNT(aa.id) FILTER (WHERE aa.status IN ('submitted','finalised')) AS completed,
				ROUND(AVG(aa.percentage) FILTER (WHERE aa.status IN ('submitted','finalised')), 2) AS avg_score,
				MIN(aa.percentage) FILTER (WHERE aa.status IN ('submitted','finalised')) AS min_score,
				MAX(aa.percentage) FILTER (WHERE aa.status IN ('submitted','finalised')) AS max_score
			FROM quizzes q
			LEFT JOIN assessment_attempts aa ON aa.quiz_id = q.id
			WHERE q.id = $1
			GROUP BY q.title, q.is_public
		`, quizID)
		var title string
		var isPublic bool
		var total, completed sql.NullInt64
		var avg, min, max sql.NullFloat64
		if err := row.Scan(&title, &isPublic, &total, &completed, &avg, &min, &max); err != nil {
			return
		}
		for _, pair := range [][2]string{
			{"Title", title},
			{"Public", boolStr(isPublic)},
			{"Total Attempts", nullInt64Str(total)},
			{"Completed Attempts", nullInt64Str(completed)},
			{"Average Score %", nullFloat64Str(avg)},
			{"Min Score %", nullFloat64Str(min)},
			{"Max Score %", nullFloat64Str(max)},
		} {
			ch <- []string{pair[0], pair[1]}
		}
	}()
	return headers, ch, nil
}

func (svc *ExportService) streamQuizAttempts(ctx context.Context, tx *goqulib.TxDatabase, quizID string, filtersJSON json.RawMessage) ([]string, <-chan []string, error) {
	headers := []string{"Attempt ID", "User ID", "Status", "Score", "Percentage", "Started At", "Submitted At", "Duration (s)"}
	ch := make(chan []string, exportBatchSize)

	var statusFilter *string
	if len(filtersJSON) > 0 {
		var f struct {
			Status *string `json:"status"`
		}
		_ = json.Unmarshal(filtersJSON, &f)
		statusFilter = f.Status
	}

	go func() {
		defer close(ch)
		query := `
			SELECT aa.id, aa.user_id, aa.status,
				aa.score, aa.percentage,
				aa.started_at, aa.submitted_at,
				EXTRACT(EPOCH FROM (aa.submitted_at - aa.started_at))::int AS duration_seconds
			FROM assessment_attempts aa
			WHERE aa.quiz_id = $1
		`
		args := []interface{}{quizID}
		if statusFilter != nil {
			query += ` AND aa.status = $2`
			args = append(args, *statusFilter)
		}
		query += ` ORDER BY aa.started_at DESC`
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			svc.logger.Error("export streamQuizAttempts", zap.Error(err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id, userID, status string
			var score, pct sql.NullFloat64
			var startedAt time.Time
			var submittedAt sql.NullTime
			var durSec sql.NullInt64
			if err := rows.Scan(&id, &userID, &status, &score, &pct, &startedAt, &submittedAt, &durSec); err != nil {
				return
			}
			ch <- []string{
				id, userID, status,
				nullFloat64Str(score), nullFloat64Str(pct),
				startedAt.Format(time.RFC3339),
				nullTimeStr(submittedAt),
				nullInt64Str(durSec),
			}
		}
	}()
	return headers, ch, nil
}

func (svc *ExportService) streamQuestionMetrics(ctx context.Context, tx *goqulib.TxDatabase, quizID string, _ json.RawMessage) ([]string, <-chan []string, error) {
	headers := []string{"Question ID", "Question Text", "Correct Attempts", "Total Attempts", "Correct Rate %", "Avg Response Time (s)"}
	ch := make(chan []string, exportBatchSize)
	go func() {
		defer close(ch)
		rows, err := tx.QueryContext(ctx, `
			SELECT
				q.id,
				LEFT(q.question, 80) AS question_text,
				COUNT(aa_a.id) FILTER (WHERE aa_a.is_correct = true) AS correct_attempts,
				COUNT(aa_a.id) AS total_attempts,
				ROUND(
					100.0 * COUNT(aa_a.id) FILTER (WHERE aa_a.is_correct = true)
					/ NULLIF(COUNT(aa_a.id), 0), 2
				) AS correct_rate,
				ROUND(AVG(aa_a.response_time), 2) AS avg_response_time
			FROM questions q
			JOIN quiz_questions qq ON qq.question_id = q.id AND qq.quiz_id = $1
			LEFT JOIN assessment_attempt_answers aa_a ON aa_a.question_id = q.id
				AND aa_a.attempt_id IN (SELECT id FROM assessment_attempts WHERE quiz_id = $1)
			GROUP BY q.id, question_text
			ORDER BY correct_rate ASC NULLS LAST
		`, quizID)
		if err != nil {
			svc.logger.Error("export streamQuestionMetrics", zap.Error(err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id, text string
			var correct, total sql.NullInt64
			var rate, avgRT sql.NullFloat64
			if err := rows.Scan(&id, &text, &correct, &total, &rate, &avgRT); err != nil {
				return
			}
			ch <- []string{id, text, nullInt64Str(correct), nullInt64Str(total), nullFloat64Str(rate), nullFloat64Str(avgRT)}
		}
	}()
	return headers, ch, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// State transitions
// ─────────────────────────────────────────────────────────────────────────────

func (svc *ExportService) markRunning(ctx context.Context, reportID uuid.UUID) error {
	_, err := svc.db.ExecContext(ctx,
		`UPDATE generated_reports SET status = 'RUNNING', started_at = now() WHERE id = $1 AND status = 'QUEUED'`,
		reportID,
	)
	return err
}

func (svc *ExportService) markCompleted(ctx context.Context, reportID uuid.UUID, storageKey, provider string, rowCount int, snapshotAt time.Time) error {
	expiresAt := time.Now().UTC().Add(time.Duration(svc.retentionDays) * 24 * time.Hour)
	_, err := svc.db.ExecContext(ctx,
		`UPDATE generated_reports SET
			status = 'COMPLETED', storage_key = $2, storage_provider = $3,
			row_count = $4, snapshot_started_at = $5, completed_at = now(), expires_at = $6
		WHERE id = $1`,
		reportID, storageKey, provider, rowCount, snapshotAt, expiresAt,
	)
	if err != nil {
		svc.logger.Error("export markCompleted failed", zap.String("report_id", reportID.String()), zap.Error(err))
	}
	svc.writeAudit(ctx, "", &reportID, "EXPORT_COMPLETED", "", "", nil, 0, rowCount, true)
	return err
}

func (svc *ExportService) markFailed(ctx context.Context, reportID uuid.UUID, errMsg string) error {
	_, err := svc.db.ExecContext(ctx,
		`UPDATE generated_reports SET status = 'FAILED', error_message = $2 WHERE id = $1`,
		reportID, truncateStr(errMsg, 500),
	)
	svc.writeAudit(ctx, "", &reportID, "EXPORT_FAILED", "", "", nil, 0, 0, false)
	return err
}

func (svc *ExportService) writeAudit(ctx context.Context, instructorID string, reportID *uuid.UUID, action, exportType, exportFormat string, filtersJSON json.RawMessage, durationMs, rowCount int, success bool) {
	var rID *string
	if reportID != nil {
		s := reportID.String()
		rID = &s
	}
	svc.db.ExecContext(ctx,
		`INSERT INTO export_audit_log
			(instructor_id, report_id, action, export_type, export_format, filters_json, duration_ms, row_count, success)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		instructorID, rID, action, exportType, exportFormat, safeFilterJSON(filtersJSON), durationMs, rowCount, success,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func storageKey(reportID uuid.UUID, ext string) string {
	return fmt.Sprintf("reports/%s.%s", reportID.String(), ext)
}

func nullInt64Str(v sql.NullInt64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatInt(v.Int64, 10)
}

func nullFloat64Str(v sql.NullFloat64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatFloat(v.Float64, 'f', 2, 64)
}

func nullTimeStr(v sql.NullTime) string {
	if !v.Valid {
		return ""
	}
	return v.Time.Format(time.RFC3339)
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}

func exportTypeLabel(et structs.ExportType) string {
	return strings.ReplaceAll(strings.Title(strings.ToLower(string(et))), "_", " ")
}

func safeFilterJSON(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
