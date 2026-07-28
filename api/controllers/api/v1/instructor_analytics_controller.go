package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	"go.uber.org/zap"
)

type InstructorAnalyticsController struct {
	svc    *services.InstructorAnalyticsService
	logger *zap.Logger
}

func NewInstructorAnalyticsController(svc *services.InstructorAnalyticsService, logger *zap.Logger) *InstructorAnalyticsController {
	return &InstructorAnalyticsController{svc: svc, logger: logger}
}

func (ctrl *InstructorAnalyticsController) Service() *services.InstructorAnalyticsService {
	return ctrl.svc
}

func (ctrl *InstructorAnalyticsController) currentUser(c *fiber.Ctx) (models.User, bool) {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	return user, ok && user.ID != ""
}

func (ctrl *InstructorAnalyticsController) preferredTimezone(c *fiber.Ctx) string {
	return strings.TrimSpace(c.Query("timezone"))
}

// GET /instructor/analytics/overview
func (ctrl *InstructorAnalyticsController) GetOverview(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	resp, err := ctrl.svc.GetPortfolioOverview(c.Context(), user.ID, ctrl.preferredTimezone(c))
	if err != nil {
		ctrl.logger.Error("instructor portfolio overview failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load instructor portfolio overview")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /instructor/analytics/quizzes
func (ctrl *InstructorAnalyticsController) GetQuizList(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")
	sortBy := c.Query("sort_by")
	sortDir := c.Query("sort_dir")

	resp, err := ctrl.svc.GetOwnedQuizList(c.Context(), user.ID, ctrl.preferredTimezone(c), cursor, limit, sortBy, sortDir)
	if err != nil {
		if errors.Is(err, services.ErrInstructorInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("instructor quiz list failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load instructor quiz list")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /instructor/analytics/learners
func (ctrl *InstructorAnalyticsController) GetLearnerList(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")
	search := c.Query("q")
	quizIDFilter := c.Query("quiz_id")
	statusFilter := c.Query("status")
	sortBy := c.Query("sort_by")
	sortDir := c.Query("sort_dir")

	resp, err := ctrl.svc.GetLearnerList(c.Context(), user.ID, ctrl.preferredTimezone(c), cursor, limit, search, quizIDFilter, statusFilter, sortBy, sortDir)
	if err != nil {
		if errors.Is(err, services.ErrInstructorInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("instructor learner list failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load learner performance list")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /instructor/analytics/releases
func (ctrl *InstructorAnalyticsController) GetReleaseMonitoring(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")
	classification := c.Query("classification")

	resp, err := ctrl.svc.GetReleaseMonitoring(c.Context(), user.ID, ctrl.preferredTimezone(c), cursor, limit, classification)
	if err != nil {
		if errors.Is(err, services.ErrInstructorInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("instructor release monitoring failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load release monitoring report")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /instructor/analytics/timeline
func (ctrl *InstructorAnalyticsController) GetTimeline(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")

	resp, err := ctrl.svc.GetTimeline(c.Context(), user.ID, ctrl.preferredTimezone(c), cursor, limit)
	if err != nil {
		if errors.Is(err, services.ErrInstructorInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("instructor timeline failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load activity timeline")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /quizzes/:quiz_id/analytics/summary
func (ctrl *InstructorAnalyticsController) GetQuizSummary(c *fiber.Ctx) error {
	_, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.QuizId)))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id")
	}
	resp, err := ctrl.svc.GetQuizCohortSummary(c.Context(), quizID, ctrl.preferredTimezone(c))
	if err != nil {
		if errors.Is(err, services.ErrInstructorQuizNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "Quiz not found")
		}
		ctrl.logger.Error("instructor quiz summary failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load quiz cohort summary")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /quizzes/:quiz_id/analytics/attempts
func (ctrl *InstructorAnalyticsController) GetQuizAttempts(c *fiber.Ctx) error {
	_, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.QuizId)))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id")
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")
	statusFilter := c.Query("status")

	resp, err := ctrl.svc.GetQuizAttemptList(c.Context(), quizID, cursor, limit, statusFilter)
	if err != nil {
		if errors.Is(err, services.ErrInstructorInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("instructor quiz attempts failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load quiz attempt list")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /quizzes/:quiz_id/analytics/questions
func (ctrl *InstructorAnalyticsController) GetQuestionMetrics(c *fiber.Ctx) error {
	_, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.QuizId)))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id")
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")

	resp, err := ctrl.svc.GetQuestionMetrics(c.Context(), quizID, cursor, limit)
	if err != nil {
		if errors.Is(err, services.ErrInstructorInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("instructor question metrics failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load question metrics")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

// GET /quizzes/:quiz_id/analytics/engagement
func (ctrl *InstructorAnalyticsController) GetEngagement(c *fiber.Ctx) error {
	_, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.QuizId)))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id")
	}
	resp, err := ctrl.svc.GetEngagementMetrics(c.Context(), quizID)
	if err != nil {
		ctrl.logger.Error("instructor engagement failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load engagement metrics")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}
