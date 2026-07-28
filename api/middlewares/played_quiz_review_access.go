package middlewares

import (
	"database/sql"
	"net/http"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func playedQuizReviewIDFromRequest(c *fiber.Ctx) string {
	if id := c.Params(constants.UserPlayedQuizId); id != "" {
		return id
	}
	return c.Query(constants.UserPlayedQuiz)
}

func isValidPlayedQuizReviewID(id string) bool {
	return id != "" && len(id) == 36
}

// VerifyPlayedQuizReviewAccess gates answer-key review endpoints to the participant,
// live session host, or an authorised quiz editor.
func (m *Middleware) VerifyPlayedQuizReviewAccess(c *fiber.Ctx) error {
	userID := quizUtilsHelper.GetString(c.Locals(constants.ContextUid))
	if userID == "" || userID == "<nil>" {
		return utils.JSONFail(c, http.StatusUnauthorized, constants.Unauthenticated)
	}

	userPlayedQuizID := playedQuizReviewIDFromRequest(c)
	if !isValidPlayedQuizReviewID(userPlayedQuizID) {
		return utils.JSONFail(c, http.StatusBadRequest, "user play quiz should be valid string")
	}

	var editor *models.User
	if user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser)); ok {
		editor = &user
	}

	allowed, err := m.userPlayedQuizModel.CanAccessReview(
		userID,
		editor,
		&m.Config,
		m.sharedQuizzesModel,
		userPlayedQuizID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.JSONFail(c, http.StatusNotFound, "user played quiz not found")
		}
		m.Logger.Error("played quiz review access check failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUnauthenticated)
	}
	if !allowed {
		return utils.JSONError(c, http.StatusForbidden, constants.ErrUnauthorized)
	}

	return c.Next()
}
