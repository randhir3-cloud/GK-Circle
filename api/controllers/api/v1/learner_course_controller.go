package v1

import (
	"errors"
	"net/http"

	"github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

// LearnerCourseController exposes published Course catalog and outline APIs.
type LearnerCourseController struct {
	courseModel     *models.CourseModel
	courseNodeModel *models.CourseNodeModel
	logger          *zap.Logger
}

func InitLearnerCourseController(db *goqu.Database, logger *zap.Logger, _ *config.AppConfig) (*LearnerCourseController, error) {
	return &LearnerCourseController{
		courseModel:     models.InitCourseModel(db),
		courseNodeModel: models.InitCourseNodeModel(db),
		logger:          logger,
	}, nil
}

func (ctrl *LearnerCourseController) ListPublishedCourses(c *fiber.Ctx) error {
	courses, err := ctrl.courseModel.ListPublishedCourses()
	if err != nil {
		ctrl.logger.Error("list published courses failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourse)
	}

	response := make([]structs.LearnerCourseResponse, 0, len(courses))
	for _, course := range courses {
		response = append(response, toLearnerCourseResponse(course))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *LearnerCourseController) GetPublishedCourse(c *fiber.Ctx) error {
	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	course, err := ctrl.courseModel.RequirePublishedCourse(courseID)
	if err != nil {
		if errors.Is(err, models.ErrCourseNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		}
		if errors.Is(err, models.ErrCourseNotPublished) {
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		}
		ctrl.logger.Error("get published course failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourse)
	}

	return utils.JSONSuccess(c, http.StatusOK, toLearnerCourseResponse(course))
}

func (ctrl *LearnerCourseController) GetPublishedOutline(c *fiber.Ctx) error {
	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	if _, err := ctrl.courseModel.RequirePublishedCourse(courseID); err != nil {
		if errors.Is(err, models.ErrCourseNotFound) || errors.Is(err, models.ErrCourseNotPublished) {
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		}
		ctrl.logger.Error("published course lookup failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourse)
	}

	hierarchy, err := ctrl.courseNodeModel.GetHierarchy(courseID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrCourseNodeCourseNotFound):
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		case errors.Is(err, models.ErrCourseNodeHierarchyIntegrity):
			ctrl.logger.Error("course hierarchy integrity failure", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourseHierarchy)
		default:
			ctrl.logger.Error("error while getting learner course outline", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourseHierarchy)
		}
	}

	return utils.JSONSuccess(c, http.StatusOK, toAdminCourseHierarchyResponse(hierarchy))
}

func toLearnerCourseResponse(course models.Course) structs.LearnerCourseResponse {
	return structs.LearnerCourseResponse{
		ID:               course.ID.String(),
		Title:            course.Title,
		ShortDescription: nullStringPtr(course.ShortDescription),
		Language:         nullStringPtr(course.Language),
		Difficulty:       nullStringPtr(course.Difficulty),
		Status:           string(course.Status),
	}
}
