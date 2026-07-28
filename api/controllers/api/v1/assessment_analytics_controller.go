package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	goqu "github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	"go.uber.org/zap"
)

type AssessmentAnalyticsController struct {
	analyticsSvc *services.AssessmentAnalyticsService
	logger       *zap.Logger
}

func NewAssessmentAnalyticsController(db *goqu.Database, logger *zap.Logger) *AssessmentAnalyticsController {
	return &AssessmentAnalyticsController{
		analyticsSvc: services.NewAssessmentAnalyticsService(db, logger),
		logger:       logger,
	}
}

func (ctrl *AssessmentAnalyticsController) SetLearnerAnalyticsCache(cache *services.LearnerAnalyticsCache) {
	if ctrl.analyticsSvc != nil {
		ctrl.analyticsSvc.SetLearnerAnalyticsCache(cache)
	}
}

func (ctrl *AssessmentAnalyticsController) RecordClientTelemetryBatch(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	var req structs.RecordTelemetryBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid request body")
	}

	correlationID := utils.ResolveAuditCorrelationID(c)
	c.Set(utils.HeaderCorrelationID, correlationID)

	result, err := ctrl.analyticsSvc.RecordClientTelemetryBatch(quizID, attemptID, user.ID, correlationID, req.Events)
	if err != nil {
		return mapAnalyticsError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, result)
}

func (ctrl *AssessmentAnalyticsController) GetAttemptEvents(c *fiber.Ctx) error {
	quizID, attemptID, err := parseQuizAndAttemptIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return utils.JSONFail(c, http.StatusBadRequest, "Invalid limit")
		}
		limit = parsed
	}

	events, nextCursor, err := ctrl.analyticsSvc.ListAttemptEvents(quizID, attemptID, user.ID, models.AnalyticsPaginationParams{
		Limit:  limit,
		Cursor: c.Query("cursor"),
	})
	if err != nil {
		return mapAnalyticsError(c, err)
	}

	responseEvents := make([]structs.AssessmentAnalyticsEventResponse, 0, len(events))
	for _, event := range events {
		responseEvents = append(responseEvents, toAnalyticsEventResponse(event))
	}
	return utils.JSONSuccess(c, http.StatusOK, structs.AssessmentAnalyticsEventListResponse{
		Events:     responseEvents,
		NextCursor: nextCursor,
	})
}

func toAnalyticsEventResponse(event models.AssessmentAnalyticsEvent) structs.AssessmentAnalyticsEventResponse {
	resp := structs.AssessmentAnalyticsEventResponse{
		ID:            event.ID.String(),
		AttemptID:     event.AttemptID.String(),
		QuizID:        event.QuizID.String(),
		UserID:        event.UserID,
		EventType:     event.EventType,
		EventSource:   event.EventSource,
		CorrelationID: event.CorrelationID,
		SchemaVersion: event.SchemaVersion,
		Metadata:      map[string]interface{}{},
		OccurredAt:    event.OccurredAt.UTC().Format(time.RFC3339),
		CreatedAt:     event.CreatedAt.UTC().Format(time.RFC3339),
	}
	if event.ClientEventID.Valid {
		value := event.ClientEventID.UUID.String()
		resp.ClientEventID = &value
	}
	if event.QuizOwnerID.Valid {
		value := event.QuizOwnerID.String
		resp.QuizOwnerID = &value
	}
	if event.IdempotencyKey.Valid {
		value := event.IdempotencyKey.String
		resp.IdempotencyKey = &value
	}
	if len(event.Metadata) > 0 {
		_ = json.Unmarshal(event.Metadata, &resp.Metadata)
		if resp.Metadata == nil {
			resp.Metadata = map[string]interface{}{}
		}
	}
	return resp
}

func mapAnalyticsError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrAssessmentAttemptNotFound):
		return utils.JSONFail(c, http.StatusNotFound, "Attempt not found")
	case errors.Is(err, services.ErrAnalyticsInvalidEventType),
		errors.Is(err, services.ErrAnalyticsInvalidOccurredAt),
		errors.Is(err, services.ErrAnalyticsInvalidMetadata),
		errors.Is(err, services.ErrAnalyticsSensitiveMetadata),
		errors.Is(err, services.ErrAnalyticsBatchEmpty),
		errors.Is(err, services.ErrAnalyticsBatchTooLarge),
		errors.Is(err, models.ErrAnalyticsInvalidCursor):
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrAnalyticsAttemptMismatch):
		return utils.JSONFail(c, http.StatusForbidden, "Not allowed to record analytics for this attempt")
	default:
		ctrlLog := zap.L()
		ctrlLog.Error("analytics request failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to process analytics request")
	}
}
