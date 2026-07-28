package v1

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

const (
	courseTitleMaxLen      = 200
	courseShortDescMaxLen  = 2000
	courseOptionalFieldMax = 64
)

type CourseController struct {
	courseModel *models.CourseModel
	appConfig   *config.AppConfig
	logger      *zap.Logger
}

func InitCourseController(db *goqu.Database, logger *zap.Logger, appConfig *config.AppConfig) (*CourseController, error) {
	return &CourseController{
		courseModel: models.InitCourseModel(db),
		appConfig:   appConfig,
		logger:      logger,
	}, nil
}

func (ctrl *CourseController) CreateCourse(c *fiber.Ctx) error {
	_, ownerID, err := authorizeCourseAdmin(c, ctrl.appConfig)
	if err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	var req structs.ReqCreateAdminCourse
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	title, shortDesc, language, difficulty, visibility, err := validateCreateAdminCourse(req)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	course, err := ctrl.courseModel.CreateCourse(models.CreateCourseParams{
		OwnerID:          ownerID,
		Title:            title,
		ShortDescription: shortDesc,
		Language:         language,
		Difficulty:       difficulty,
		Visibility:       visibility,
	})
	if err != nil {
		if errors.Is(err, models.ErrCourseTitleRequired) || errors.Is(err, models.ErrCourseOwnerRequired) {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
		ctrl.logger.Error("error while creating course", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrCreateCourse)
	}
	return utils.JSONSuccess(c, http.StatusCreated, toAdminCourseResponse(course))
}

func (ctrl *CourseController) ListCourses(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courses, err := ctrl.courseModel.ListCourses()
	if err != nil {
		ctrl.logger.Error("error while listing courses", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListCourses)
	}
	response := make([]structs.AdminCourseResponse, 0, len(courses))
	for _, course := range courses {
		response = append(response, toAdminCourseResponse(course))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *CourseController) GetCourse(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	course, err := ctrl.courseModel.GetCourseByID(courseID)
	if err != nil {
		if errors.Is(err, models.ErrCourseNotFound) {
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		}
		ctrl.logger.Error("error while getting course", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetCourse)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAdminCourseResponse(course))
}

func (ctrl *CourseController) UpdateCourse(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}

	var req structs.ReqUpdateAdminCourse
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}
	params, err := validateUpdateAdminCourse(req)
	if err != nil {
		if errors.Is(err, models.ErrCourseUpdateRequired) {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseEmptyPatch)
		}
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	course, err := ctrl.courseModel.UpdateCourse(courseID, params)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrCourseUpdateRequired):
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseEmptyPatch)
		case errors.Is(err, models.ErrCourseTitleRequired):
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseTitleInvalid)
		case errors.Is(err, models.ErrCourseStatusInvalid):
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseStatusInvalid)
		case errors.Is(err, models.ErrCourseNotFound):
			return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
		default:
			ctrl.logger.Error("error while updating course", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUpdateCourse)
		}
	}
	return utils.JSONSuccess(c, http.StatusOK, toAdminCourseResponse(course))
}

func parseCourseIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(c.Params(constants.CourseId)))
}

func validateCreateAdminCourse(req structs.ReqCreateAdminCourse) (title, shortDesc, language, difficulty, visibility string, err error) {
	title = strings.TrimSpace(req.Title)
	if title == "" || utf8.RuneCountInString(title) > courseTitleMaxLen {
		return "", "", "", "", "", errors.New(constants.ErrCourseTitleInvalid)
	}
	shortDesc, err = optionalCreateString(req.ShortDescription, courseShortDescMaxLen)
	if err != nil {
		return "", "", "", "", "", err
	}
	language, err = optionalCreateString(req.Language, courseOptionalFieldMax)
	if err != nil {
		return "", "", "", "", "", err
	}
	difficulty, err = optionalCreateString(req.Difficulty, courseOptionalFieldMax)
	if err != nil {
		return "", "", "", "", "", err
	}
	visibility, err = optionalCreateString(req.Visibility, courseOptionalFieldMax)
	if err != nil {
		return "", "", "", "", "", err
	}
	return title, shortDesc, language, difficulty, visibility, nil
}

func validateUpdateAdminCourse(req structs.ReqUpdateAdminCourse) (models.UpdateCourseParams, error) {
	params := models.UpdateCourseParams{}
	supplied := false

	if req.Title.Present {
		supplied = true
		if req.Title.Null {
			return params, errors.New(constants.ErrCourseTitleInvalid)
		}
		title := strings.TrimSpace(req.Title.Value)
		if title == "" || utf8.RuneCountInString(title) > courseTitleMaxLen {
			return params, errors.New(constants.ErrCourseTitleInvalid)
		}
		params.Title = &title
	}
	var err error
	if params.ShortDescription, supplied, err = optionalUpdateField(req.ShortDescription, courseShortDescMaxLen, supplied); err != nil {
		return params, err
	}
	if params.Language, supplied, err = optionalUpdateField(req.Language, courseOptionalFieldMax, supplied); err != nil {
		return params, err
	}
	if params.Difficulty, supplied, err = optionalUpdateField(req.Difficulty, courseOptionalFieldMax, supplied); err != nil {
		return params, err
	}
	if params.Visibility, supplied, err = optionalUpdateField(req.Visibility, courseOptionalFieldMax, supplied); err != nil {
		return params, err
	}
	if req.Status.Present {
		supplied = true
		if req.Status.Null {
			return params, errors.New(constants.ErrCourseStatusInvalid)
		}
		status := models.CourseStatus(strings.TrimSpace(req.Status.Value))
		switch status {
		case models.CourseStatusDraft, models.CourseStatusPublished, models.CourseStatusArchived:
			params.Status = &status
		default:
			return params, errors.New(constants.ErrCourseStatusInvalid)
		}
	}
	if !supplied {
		return params, models.ErrCourseUpdateRequired
	}
	return params, nil
}

func optionalCreateString(field structs.OptionalString, maxLen int) (string, error) {
	if !field.Present || field.Null {
		return "", nil
	}
	value := strings.TrimSpace(field.Value)
	if utf8.RuneCountInString(value) > maxLen {
		return "", errors.New(constants.ErrCourseFieldTooLong)
	}
	return value, nil
}

func optionalUpdateField(field structs.OptionalString, maxLen int, supplied bool) (models.OptionalNullableString, bool, error) {
	if !field.Present {
		return models.OptionalNullableString{}, supplied, nil
	}
	if field.Null {
		return models.OptionalNullableString{Present: true, Null: true}, true, nil
	}
	value := strings.TrimSpace(field.Value)
	if utf8.RuneCountInString(value) > maxLen {
		return models.OptionalNullableString{}, supplied, errors.New(constants.ErrCourseFieldTooLong)
	}
	return models.OptionalNullableString{Present: true, Value: value}, true, nil
}

func toAdminCourseResponse(course models.Course) structs.AdminCourseResponse {
	return structs.AdminCourseResponse{
		ID:               course.ID.String(),
		OwnerID:          course.OwnerID,
		Title:            course.Title,
		ShortDescription: nullStringPtr(course.ShortDescription),
		Language:         nullStringPtr(course.Language),
		Difficulty:       nullStringPtr(course.Difficulty),
		Visibility:       nullStringPtr(course.Visibility),
		Status:           string(course.Status),
		CreatedAt:        course.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:        course.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
