package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	"go.uber.org/zap"
)

// InstructorExportController handles export job creation, scheduling, history, and downloads.
// Authorization model:
//   - Portfolio endpoints: authenticated instructor only (owner).
//   - Per-quiz endpoints: collaborators allowed via VerifyQuizAnalyticsAccess middleware (T03 pattern).
//   - Download: explicit ownership check — report.instructor_id == session.user_id (not UUID secrecy).
type InstructorExportController struct {
	exportSvc *services.ExportService
	workerCh  chan<- uuid.UUID
	storage   services.StorageProvider
	logger    *zap.Logger
}

// NewInstructorExportController creates the controller.
func NewInstructorExportController(
	exportSvc *services.ExportService,
	workerCh chan<- uuid.UUID,
	storage services.StorageProvider,
	logger *zap.Logger,
) *InstructorExportController {
	return &InstructorExportController{
		exportSvc: exportSvc,
		workerCh:  workerCh,
		storage:   storage,
		logger:    logger,
	}
}

func (ctrl *InstructorExportController) currentUser(c *fiber.Ctx) (models.User, bool) {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	return user, ok && user.ID != ""
}

// ─────────────────────────────────────────────────────────────────────────────
// Portfolio-scope: one-time exports
// ─────────────────────────────────────────────────────────────────────────────

// RequestExport handles POST /instructor/reports/exports. Returns 202 Accepted + job ID.
func (ctrl *InstructorExportController) RequestExport(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	var body structs.RequestExportInput
	if err := c.BodyParser(&body); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validateExportType(body.ExportType); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	if err := validateExportFormat(body.ExportFormat); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	if body.Title == "" {
		body.Title = fmt.Sprintf("%s %s Export", body.ExportType, body.ExportFormat)
	}

	filtersRaw, _ := json.Marshal(body.FiltersJSON)
	jobID := uuid.New()

	if err := ctrl.exportSvc.InsertJob(c.Context(), jobID, user.ID, body.Title, body.ExportType, body.ExportFormat, filtersRaw, body.QuizID); err != nil {
		ctrl.logger.Error("request export: insert job", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "failed to queue export job")
	}

	select {
	case ctrl.workerCh <- jobID:
	default:
	}

	c.Status(http.StatusAccepted)
	return utils.JSONSuccess(c, http.StatusAccepted, map[string]string{"id": jobID.String(), "status": "QUEUED"})
}

// GetExportStatus handles GET /instructor/reports/exports/:report_id.
func (ctrl *InstructorExportController) GetExportStatus(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	reportID, err := uuid.Parse(c.Params("report_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid report_id")
	}

	job, err := ctrl.exportSvc.GetJob(c.Context(), reportID, user.ID)
	if errors.Is(err, services.ErrReportNotFound) {
		return utils.JSONFail(c, http.StatusNotFound, "report not found")
	}
	if errors.Is(err, services.ErrReportUnauthorized) {
		return utils.JSONFail(c, http.StatusForbidden, "access denied")
	}
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to load report")
	}
	return utils.JSONSuccess(c, http.StatusOK, job)
}

// DownloadReport handles GET /instructor/reports/exports/:report_id/download.
// Explicit ownership: report.instructor_id == session.user_id (not UUID secrecy).
func (ctrl *InstructorExportController) DownloadReport(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	reportID, err := uuid.Parse(c.Params("report_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid report_id")
	}

	meta, err := ctrl.exportSvc.GetDownloadMeta(c.Context(), reportID, user.ID)
	if errors.Is(err, services.ErrReportNotFound) {
		return utils.JSONFail(c, http.StatusNotFound, "report not found")
	}
	if errors.Is(err, services.ErrReportUnauthorized) {
		return utils.JSONFail(c, http.StatusForbidden, "access denied")
	}
	if errors.Is(err, services.ErrReportDeleted) {
		return utils.JSONFail(c, http.StatusGone, "report has been deleted")
	}
	if errors.Is(err, services.ErrReportNotCompleted) {
		return utils.JSONFail(c, http.StatusNotFound, "report is not yet completed")
	}
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to load report")
	}

	rc, size, err := ctrl.storage.Get(c.Context(), meta.StorageKey)
	if err != nil {
		ctrl.logger.Error("download: storage get", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "storage unavailable")
	}
	defer rc.Close()

	ext := formatExt(meta.ExportFormat)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="report.%s"`, ext))
	c.Set("Content-Type", services.ContentTypeForFormat(meta.ExportFormat))
	if size > 0 {
		c.Set("Content-Length", strconv.FormatInt(size, 10))
	}

	// Record download — non-fatal.
	go ctrl.exportSvc.RecordDownload(c.Context(), reportID, user.ID, c.IP(), c.Get("User-Agent"))

	return c.SendStream(rc)
}

// DeleteReport handles DELETE /instructor/reports/exports/:report_id.
// QUEUED/RUNNING → CANCELLED. COMPLETED → soft-delete (storage purged, deleted_at set, row kept).
func (ctrl *InstructorExportController) DeleteReport(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	reportID, err := uuid.Parse(c.Params("report_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid report_id")
	}
	if err := ctrl.exportSvc.DeleteReport(c.Context(), reportID, user.ID, ctrl.storage); err != nil {
		if errors.Is(err, services.ErrReportNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "report not found")
		}
		if errors.Is(err, services.ErrReportUnauthorized) {
			return utils.JSONFail(c, http.StatusForbidden, "access denied")
		}
		return utils.JSONError(c, http.StatusInternalServerError, "failed to delete report")
	}
	c.Status(http.StatusNoContent)
	return nil
}

// GetHistory handles GET /instructor/reports/history.
func (ctrl *InstructorExportController) GetHistory(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	cursor := c.Query("cursor")
	limit := queryIntDefault(c, "limit", 20, 100)

	result, err := ctrl.exportSvc.ListHistory(c.Context(), user.ID, cursor, limit)
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to list reports")
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

// GetAuditLog handles GET /instructor/reports/audit.
func (ctrl *InstructorExportController) GetAuditLog(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	cursor := c.Query("cursor")
	limit := queryIntDefault(c, "limit", 20, 100)

	result, err := ctrl.exportSvc.ListAuditLog(c.Context(), user.ID, cursor, limit)
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to list audit log")
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled reports
// ─────────────────────────────────────────────────────────────────────────────

// CreateSchedule handles POST /instructor/reports/schedules.
// Validates cron_expr at request time — 400 on invalid before any DB write.
func (ctrl *InstructorExportController) CreateSchedule(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	var body structs.CreateScheduleInput
	if err := c.BodyParser(&body); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validateCreateSchedule(body); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	result, err := ctrl.exportSvc.CreateSchedule(c.Context(), user.ID, body)
	if err != nil {
		ctrl.logger.Error("create schedule", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "failed to create schedule")
	}
	return utils.JSONSuccess(c, http.StatusCreated, result)
}

// ListSchedules handles GET /instructor/reports/schedules.
func (ctrl *InstructorExportController) ListSchedules(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	result, err := ctrl.exportSvc.ListSchedules(c.Context(), user.ID)
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to list schedules")
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

// GetSchedule handles GET /instructor/reports/schedules/:schedule_id.
func (ctrl *InstructorExportController) GetSchedule(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid schedule_id")
	}
	result, err := ctrl.exportSvc.GetSchedule(c.Context(), scheduleID, user.ID)
	if errors.Is(err, services.ErrScheduleNotFound) {
		return utils.JSONFail(c, http.StatusNotFound, "schedule not found")
	}
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to get schedule")
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

// UpdateSchedule handles PATCH /instructor/reports/schedules/:schedule_id.
func (ctrl *InstructorExportController) UpdateSchedule(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid schedule_id")
	}
	var body structs.UpdateScheduleInput
	if err := c.BodyParser(&body); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid request body")
	}
	if body.CronExpr != nil && *body.CronExpr != "" {
		if err := services.ValidateCronExpr(*body.CronExpr); err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, "invalid cron_expr: "+err.Error())
		}
	}
	if body.Timezone != nil {
		if err := services.ValidateIANATimezone(*body.Timezone); err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
	}

	result, err := ctrl.exportSvc.UpdateSchedule(c.Context(), scheduleID, user.ID, body)
	if errors.Is(err, services.ErrScheduleNotFound) {
		return utils.JSONFail(c, http.StatusNotFound, "schedule not found")
	}
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to update schedule")
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

// DeleteSchedule handles DELETE /instructor/reports/schedules/:schedule_id.
func (ctrl *InstructorExportController) DeleteSchedule(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid schedule_id")
	}
	if err := ctrl.exportSvc.DeleteSchedule(c.Context(), scheduleID, user.ID); err != nil {
		if errors.Is(err, services.ErrScheduleNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "schedule not found")
		}
		return utils.JSONError(c, http.StatusInternalServerError, "failed to delete schedule")
	}
	c.Status(http.StatusNoContent)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Per-quiz-scope exports (collaborators allowed via VerifyQuizAnalyticsAccess)
// ─────────────────────────────────────────────────────────────────────────────

// RequestQuizExport handles POST /quizzes/:quiz_id/reports/exports.
func (ctrl *InstructorExportController) RequestQuizExport(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizIDStr := c.Params(string(constants.QuizId))
	if _, err := uuid.Parse(quizIDStr); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid quiz_id")
	}

	var body structs.RequestExportInput
	if err := c.BodyParser(&body); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validateExportFormat(body.ExportFormat); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	if err := validateExportType(body.ExportType); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	body.QuizID = &quizIDStr
	if body.Title == "" {
		body.Title = fmt.Sprintf("Quiz %s — %s %s", quizIDStr, body.ExportType, body.ExportFormat)
	}
	filtersRaw, _ := json.Marshal(body.FiltersJSON)
	jobID := uuid.New()

	if err := ctrl.exportSvc.InsertJob(c.Context(), jobID, user.ID, body.Title, body.ExportType, body.ExportFormat, filtersRaw, &quizIDStr); err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to queue export job")
	}
	select {
	case ctrl.workerCh <- jobID:
	default:
	}
	c.Status(http.StatusAccepted)
	return utils.JSONSuccess(c, http.StatusAccepted, map[string]string{"id": jobID.String(), "status": "QUEUED"})
}

// GetQuizHistory handles GET /quizzes/:quiz_id/reports/history.
func (ctrl *InstructorExportController) GetQuizHistory(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizIDStr := c.Params(string(constants.QuizId))
	cursor := c.Query("cursor")
	limit := queryIntDefault(c, "limit", 20, 100)

	result, err := ctrl.exportSvc.ListQuizHistory(c.Context(), user.ID, quizIDStr, cursor, limit)
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to list quiz reports")
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

// DeleteQuizReport handles DELETE /quizzes/:quiz_id/reports/exports/:report_id.
func (ctrl *InstructorExportController) DeleteQuizReport(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	reportID, err := uuid.Parse(c.Params("report_id"))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "invalid report_id")
	}
	if err := ctrl.exportSvc.DeleteReport(c.Context(), reportID, user.ID, ctrl.storage); err != nil {
		if errors.Is(err, services.ErrReportNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "report not found")
		}
		if errors.Is(err, services.ErrReportUnauthorized) {
			return utils.JSONFail(c, http.StatusForbidden, "access denied")
		}
		return utils.JSONError(c, http.StatusInternalServerError, "failed to delete report")
	}
	c.Status(http.StatusNoContent)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Validation helpers
// ─────────────────────────────────────────────────────────────────────────────

var validExportTypes = map[structs.ExportType]bool{
	structs.ExportTypePortfolioOverview:  true,
	structs.ExportTypeQuizList:           true,
	structs.ExportTypeLearnerPerformance: true,
	structs.ExportTypeReleaseMonitoring:  true,
	structs.ExportTypeTimeline:           true,
	structs.ExportTypeQuizSummary:        true,
	structs.ExportTypeQuizAttempts:       true,
	structs.ExportTypeQuestionMetrics:    true,
	structs.ExportTypeEngagement:         true,
	structs.ExportTypeFullDashboard:      true,
}

var validExportFormats = map[structs.ExportFormat]bool{
	structs.ExportFormatCSV:  true,
	structs.ExportFormatXLSX: true,
	structs.ExportFormatPDF:  true,
}

func validateExportType(t structs.ExportType) error {
	if !validExportTypes[t] {
		return fmt.Errorf("invalid export_type %q", t)
	}
	return nil
}

func validateExportFormat(f structs.ExportFormat) error {
	if !validExportFormats[f] {
		return fmt.Errorf("invalid export_format %q", f)
	}
	return nil
}

func validateCreateSchedule(body structs.CreateScheduleInput) error {
	if err := validateExportType(body.ExportType); err != nil {
		return err
	}
	if err := validateExportFormat(body.ExportFormat); err != nil {
		return err
	}
	if err := services.ValidateIANATimezone(body.Timezone); err != nil {
		return err
	}
	if body.ScheduleType != structs.ScheduleTypeOneTime {
		if body.CronExpr == nil || *body.CronExpr == "" {
			return fmt.Errorf("cron_expr is required for schedule_type %s", body.ScheduleType)
		}
		if err := services.ValidateCronExpr(*body.CronExpr); err != nil {
			return fmt.Errorf("invalid cron_expr: %s", err.Error())
		}
	}
	if body.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

func queryIntDefault(c *fiber.Ctx, key string, def, maximum int) int {
	v := c.QueryInt(key, def)
	if v <= 0 {
		v = def
	}
	if v > maximum {
		v = maximum
	}
	return v
}

func formatExt(format string) string {
	switch structs.ExportFormat(format) {
	case structs.ExportFormatCSV:
		return "csv"
	case structs.ExportFormatXLSX:
		return "xlsx"
	case structs.ExportFormatPDF:
		return "pdf"
	default:
		return "bin"
	}
}
