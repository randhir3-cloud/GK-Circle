package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

// LearnerCourseEnrollmentController exposes authenticated learner self-enrollment.
type LearnerCourseEnrollmentController struct {
	enrollmentModel *models.CourseEnrollmentModel
	logger          *zap.Logger
}

func InitLearnerCourseEnrollmentController(db *goqu.Database, logger *zap.Logger, _ *config.AppConfig) (*LearnerCourseEnrollmentController, error) {
	return &LearnerCourseEnrollmentController{
		enrollmentModel: models.InitCourseEnrollmentModel(db),
		logger:          logger,
	}, nil
}

func (ctrl *LearnerCourseEnrollmentController) GetEnrollment(c *fiber.Ctx) error {
	user, err := requireAuthenticatedLearner(c)
	if err != nil {
		return mapLearnerAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	enrolled, err := ctrl.enrollmentModel.IsUserEnrolled(user.ID, courseID)
	if err != nil {
		ctrl.logger.Error("course enrollment lookup failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourseEnrollment)
	}

	return utils.JSONSuccess(c, http.StatusOK, structs.LearnerCourseEnrollmentResponse{
		CourseID: courseID.String(),
		Enrolled: enrolled,
	})
}

func (ctrl *LearnerCourseEnrollmentController) Enroll(c *fiber.Ctx) error {
	user, err := requireAuthenticatedLearner(c)
	if err != nil {
		return mapLearnerAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	enrollment, err := ctrl.enrollmentModel.EnrollUser(user.ID, courseID)
	if err != nil {
		return ctrl.mapEnrollmentError(c, err, constants.ErrCreateCourseEnrollment)
	}

	enrolledAt := enrollment.EnrolledAt.UTC().Format(time.RFC3339)
	return utils.JSONSuccess(c, http.StatusOK, structs.LearnerCourseEnrollmentResponse{
		CourseID:   enrollment.CourseID.String(),
		Enrolled:   true,
		EnrolledAt: &enrolledAt,
	})
}

func (ctrl *LearnerCourseEnrollmentController) Unenroll(c *fiber.Ctx) error {
	user, err := requireAuthenticatedLearner(c)
	if err != nil {
		return mapLearnerAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	if err := ctrl.enrollmentModel.UnenrollUser(user.ID, courseID); err != nil {
		return ctrl.mapEnrollmentError(c, err, constants.ErrDeleteCourseEnrollment)
	}

	return utils.JSONSuccess(c, http.StatusOK, structs.LearnerCourseEnrollmentResponse{
		CourseID: courseID.String(),
		Enrolled: false,
	})
}

func (ctrl *LearnerCourseEnrollmentController) mapEnrollmentError(c *fiber.Ctx, err error, internalMessage string) error {
	switch {
	case errors.Is(err, models.ErrCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrCourseNotPublished):
		return utils.JSONFail(c, http.StatusForbidden, constants.ErrCourseNotPublished)
	case errors.Is(err, models.ErrCourseEnrollmentUserRequired):
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	default:
		ctrl.logger.Error("course enrollment request failed", zap.Error(err), zap.String("message", internalMessage))
		return utils.JSONError(c, http.StatusInternalServerError, internalMessage)
	}
}
