package middlewares

import (
	"database/sql"
	"net/http"
	"strings"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// check for read permission of perticular quiz and also save permission in context
func (m *Middleware) QuizPermission(c *fiber.Ctx) error {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok {
		m.Logger.Error("error while fetching user context from connection")
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}

	quizId := c.Params(constants.QuizId)
	if quizId == "" {
		return utils.JSONError(c, http.StatusBadRequest, "No quiz_id found")
	}

	isQuizCreator, err := m.sharedQuizzesModel.CheckQuizCreatorExists(quizId, user.ID)
	if err != nil {
		m.Logger.Error(constants.ErrGetUser, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrCheckQuizCreatorExists)
	}

	if isQuizCreator {
		c.Locals(constants.ContextQuizPermission, constants.SharePermission)
		return c.Next()
	}

	// Allowlisted public-quiz admins can fully manage any public quiz,
	// regardless of who created it (or even if the creator is gone).
	if m.Config.Quiz.IsPublicQuizAdmin(user.Email) {
		isPublic, err := m.sharedQuizzesModel.IsQuizPublic(quizId)
		if err != nil {
			m.Logger.Error(constants.ErrGetQuizPermission, zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetQuizPermission)
		}
		if isPublic {
			c.Locals(constants.ContextQuizPermission, constants.SharePermission)
			return c.Next()
		}
	}

	permission, err := m.sharedQuizzesModel.GetPermissionByQuizAndUser(quizId, user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			m.Logger.Error(constants.ErrNotAllowed, zap.Error(err))
			return utils.JSONError(c, http.StatusUnauthorized, constants.ErrNotAllowed)
		}
		m.Logger.Error(constants.ErrGetQuizPermission, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetQuizPermission)
	}

	c.Locals(constants.ContextQuizPermission, permission)
	return c.Next()
}

// check for edit permission of perticular quiz
func (m *Middleware) VerifyQuizEditAccess(c *fiber.Ctx) error {
	// Retrieve the user's permission for the current quiz from context
	permission := c.Locals(constants.ContextQuizPermission).(string)
	if permission != constants.SharePermission && permission != constants.WritePermission {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthorized)
	}

	return c.Next()
}

// check for share permission of perticular quiz
func (m *Middleware) VerifyQuizShareAccess(c *fiber.Ctx) error {
	// Retrieve the user's permission for the current quiz from context
	permission := c.Locals(constants.ContextQuizPermission).(string)
	if permission != constants.SharePermission {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthorized)
	}

	return c.Next()
}

// LoadQuizAnalyticsContext resolves owner/collaborator context without enforcing the
// legacy public-admin behavior used by QuizPermission.
func (m *Middleware) LoadQuizAnalyticsContext(c *fiber.Ctx) error {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	}
	quizID := strings.TrimSpace(c.Params(constants.QuizId))
	if quizID == "" {
		return utils.JSONError(c, http.StatusBadRequest, "No quiz_id found")
	}
	isCreator, err := m.sharedQuizzesModel.CheckQuizCreatorExists(quizID, user.ID)
	if err != nil {
		m.Logger.Error(constants.ErrCheckQuizCreatorExists, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrCheckQuizCreatorExists)
	}
	if isCreator {
		c.Locals(constants.ContextQuizID, quizID)
		c.Locals(constants.ContextQuizOwnerID, user.ID)
		c.Locals(constants.ContextQuizPermission, constants.SharePermission)
		return c.Next()
	}
	permission, err := m.sharedQuizzesModel.GetPermissionByQuizAndUser(quizID, user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Locals(constants.ContextQuizID, quizID)
			c.Locals(constants.ContextQuizOwnerID, "")
			c.Locals(constants.ContextQuizPermission, "")
			return c.Next()
		}
		m.Logger.Error(constants.ErrGetQuizPermission, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetQuizPermission)
	}
	ownerID := ""
	var ownerRow struct {
		CreatorID sql.NullString `db:"creator_id"`
	}
	found, ownerErr := m.Db.From("quizzes").Select("creator_id").Where(goqu.Ex{"id": quizID}).Limit(1).ScanStruct(&ownerRow)
	if ownerErr == nil && found && ownerRow.CreatorID.Valid {
		ownerID = ownerRow.CreatorID.String
	}
	c.Locals(constants.ContextQuizID, quizID)
	c.Locals(constants.ContextQuizOwnerID, ownerID)
	c.Locals(constants.ContextQuizPermission, permission)
	return c.Next()
}

func (m *Middleware) VerifyQuizAnalyticsAccess(c *fiber.Ctx) error {
	permission, _ := c.Locals(constants.ContextQuizPermission).(string)
	if permission != constants.SharePermission && permission != constants.WritePermission {
		return utils.JSONFail(c, http.StatusForbidden, constants.ErrUnauthorized)
	}
	return c.Next()
}
