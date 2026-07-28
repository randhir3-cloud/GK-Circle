package v1

import (
	"errors"
	"net/http"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	"go.uber.org/zap"
)

type QuizResultAdminController struct {
	adminService *services.QuizResultAdminService
	logger       *zap.Logger
}

func NewQuizResultAdminController(adminService *services.QuizResultAdminService, logger *zap.Logger) *QuizResultAdminController {
	return &QuizResultAdminController{
		adminService: adminService,
		logger:       logger,
	}
}

func (ctrl *QuizResultAdminController) GetReleaseStatus(c *fiber.Ctx) error {
	quizIDParam := c.Params(constants.QuizId)
	quizID, err := uuid.Parse(quizIDParam)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id format")
	}

	status, err := ctrl.adminService.GetReleaseStatus(quizID, utils.ResolveAuditCorrelationID(c))
	if err != nil {
		if errors.Is(err, services.ErrQuizNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "Quiz not found")
		}
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to retrieve release status")
	}

	return utils.JSONSuccess(c, http.StatusOK, status)
}

func (ctrl *QuizResultAdminController) UpdateResultSettings(c *fiber.Ctx) error {
	quizIDParam := c.Params(constants.QuizId)
	quizID, err := uuid.Parse(quizIDParam)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id format")
	}

	user, ok := c.Locals(constants.ContextUser).(models.User)
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, "Authentication required")
	}

	var req structs.UpdateQuizResultSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid request body")
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	correlationID := utils.ResolveAuditCorrelationID(c)
	c.Set(utils.HeaderCorrelationID, correlationID)

	status, err := ctrl.adminService.UpdateResultSettings(quizID, user.ID, req, ipAddress, userAgent, correlationID)
	if err != nil {
		if errors.Is(err, services.ErrQuizNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "Quiz not found")
		}
		if errors.Is(err, services.ErrInvalidReleasePolicy) || errors.Is(err, services.ErrScheduledDateRequired) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to update quiz result settings")
	}

	return utils.JSONSuccess(c, http.StatusOK, status)
}

func (ctrl *QuizResultAdminController) ReleaseResults(c *fiber.Ctx) error {
	quizIDParam := c.Params(constants.QuizId)
	quizID, err := uuid.Parse(quizIDParam)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, "Invalid quiz_id format")
	}

	user, ok := c.Locals(constants.ContextUser).(models.User)
	if !ok || user.ID == "" {
		return utils.JSONError(c, http.StatusUnauthorized, "Authentication required")
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	correlationID := utils.ResolveAuditCorrelationID(c)
	c.Set(utils.HeaderCorrelationID, correlationID)

	status, err := ctrl.adminService.ReleaseResults(quizID, user.ID, ipAddress, userAgent, correlationID)
	if err != nil {
		if errors.Is(err, services.ErrQuizNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, "Quiz not found")
		}
		if errors.Is(err, services.ErrManualReleaseNotPermitted) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		return utils.JSONError(c, http.StatusInternalServerError, "Failed to release quiz results")
	}

	return utils.JSONSuccess(c, http.StatusOK, status)
}
