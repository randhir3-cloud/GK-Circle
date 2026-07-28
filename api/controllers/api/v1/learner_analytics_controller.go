package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type LearnerAnalyticsController struct {
	svc    *services.LearnerAnalyticsAggregationService
	logger *zap.Logger
}

func NewLearnerAnalyticsController(
	db *goqu.Database,
	redisClient *goredis.Client,
	logger *zap.Logger,
) *LearnerAnalyticsController {
	cache := services.NewLearnerAnalyticsCache(redisClient, logger)
	return &LearnerAnalyticsController{
		svc:    services.NewLearnerAnalyticsAggregationService(db, cache, logger),
		logger: logger,
	}
}

func NewLearnerAnalyticsControllerWithService(
	svc *services.LearnerAnalyticsAggregationService,
	logger *zap.Logger,
) *LearnerAnalyticsController {
	return &LearnerAnalyticsController{svc: svc, logger: logger}
}

func (ctrl *LearnerAnalyticsController) Service() *services.LearnerAnalyticsAggregationService {
	return ctrl.svc
}

func (ctrl *LearnerAnalyticsController) preferredTimezone(c *fiber.Ctx) string {
	return strings.TrimSpace(c.Query("timezone"))
}

func (ctrl *LearnerAnalyticsController) currentUser(c *fiber.Ctx) (models.User, bool) {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	return user, ok && user.ID != ""
}

func (ctrl *LearnerAnalyticsController) GetDashboard(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	resp, err := ctrl.svc.GetDashboardSummary(c.Context(), user.ID, ctrl.preferredTimezone(c))
	if err != nil {
		ctrl.logger.Error("learner dashboard failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load analytics dashboard")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

func (ctrl *LearnerAnalyticsController) GetRecentActivity(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return utils.JSONFail(c, http.StatusBadRequest, "Invalid limit")
		}
		limit = parsed
	}
	resp, err := ctrl.svc.GetRecentActivity(c.Context(), user.ID, ctrl.preferredTimezone(c), c.Query("cursor"), limit)
	if err != nil {
		if errors.Is(err, services.ErrLearnerAnalyticsInvalidCursor) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("learner activity failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load recent activity")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

func (ctrl *LearnerAnalyticsController) GetPerformanceTrends(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	granularity := c.Query("granularity", "daily")
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -29)
	to := now
	if raw := c.Query("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, "Invalid from timestamp")
		}
		from = parsed.UTC()
	}
	if raw := c.Query("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, "Invalid to timestamp")
		}
		to = parsed.UTC()
	}
	resp, err := ctrl.svc.GetPerformanceTrends(c.Context(), user.ID, ctrl.preferredTimezone(c), granularity, from, to)
	if err != nil {
		if errors.Is(err, services.ErrLearnerAnalyticsInvalidRange) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("learner trends failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load performance trends")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

func (ctrl *LearnerAnalyticsController) GetSubjectPerformance(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	resp, err := ctrl.svc.GetSubjectPerformance(c.Context(), user.ID, ctrl.preferredTimezone(c))
	if err != nil {
		ctrl.logger.Error("learner subjects failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load subject performance")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}

func (ctrl *LearnerAnalyticsController) GetAttemptTimeline(c *fiber.Ctx) error {
	user, ok := ctrl.currentUser(c)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	attemptID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.AttemptId)))
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid attempt_id")
	}
	resp, err := ctrl.svc.GetAttemptTimeline(c.Context(), user.ID, ctrl.preferredTimezone(c), attemptID)
	if err != nil {
		if errors.Is(err, services.ErrLearnerAnalyticsForbidden) {
			return utils.JSONFail(c, http.StatusForbidden, "Not allowed to view this attempt timeline")
		}
		if errors.Is(err, services.ErrLearnerAnalyticsNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "Attempt not found")
		}
		ctrl.logger.Error("learner timeline failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to load attempt timeline")
	}
	return utils.JSONSuccess(c, http.StatusOK, resp)
}
