package v1

import (
	"errors"
	"net/http"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

var (
	errCourseIdentityMissing = errors.New("course identity missing")
	errCourseAdminForbidden  = errors.New("course admin forbidden")
)

func authorizeCourseAdmin(c *fiber.Ctx, appConfig *config.AppConfig) (models.User, string, error) {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || strings.TrimSpace(user.Email) == "" {
		return models.User{}, "", errCourseIdentityMissing
	}
	ownerID := strings.TrimSpace(quizUtilsHelper.GetString(c.Locals(constants.ContextUid)))
	if ownerID == "" {
		ownerID = strings.TrimSpace(user.ID)
	}
	if ownerID == "" {
		return models.User{}, "", errCourseIdentityMissing
	}
	if !models.CanManageCourses(&user) {
		return models.User{}, "", errCourseAdminForbidden
	}
	return user, ownerID, nil
}

func mapCourseAuthError(c *fiber.Ctx, logger *zap.Logger, err error) error {
	switch {
	case errors.Is(err, errCourseIdentityMissing):
		logger.Error("authenticated course administrator identity missing")
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrCourseIdentityMissing)
	case errors.Is(err, errCourseAdminForbidden):
		return utils.JSONFail(c, http.StatusForbidden, constants.ErrCourseAdminForbidden)
	default:
		logger.Error("unexpected course auth error", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUnauthenticated)
	}
}
