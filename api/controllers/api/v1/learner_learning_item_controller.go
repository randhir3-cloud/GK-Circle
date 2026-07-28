package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

// LearnerLearningItemController exposes read-only published LearningItem APIs.
// Publish filtering and learner visibility projection are repository-owned;
// this controller never filters publish_state and never iterates blocks for
// visibility. Course enrollment is enforced server-side before item delivery.
type LearnerLearningItemController struct {
	learningItemModel *models.LearningItemModel
	enrollmentModel   *models.CourseEnrollmentModel
	logger            *zap.Logger
}

func InitLearnerLearningItemController(db *goqu.Database, logger *zap.Logger, _ *config.AppConfig) (*LearnerLearningItemController, error) {
	return &LearnerLearningItemController{
		learningItemModel: models.InitLearningItemModel(db),
		enrollmentModel:   models.InitCourseEnrollmentModel(db),
		logger:            logger,
	}, nil
}

func (ctrl *LearnerLearningItemController) List(c *fiber.Ctx) error {
	user, err := requireAuthenticatedLearner(c)
	if err != nil {
		return mapLearnerAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, err := parseCourseAndNodeIDParams(c)
	if err != nil {
		return err
	}

	if err := ctrl.requireCourseEnrollment(user, courseID); err != nil {
		return ctrl.mapLearnerLearningItemError(c, err, constants.ErrListLearningItems)
	}

	items, err := ctrl.learningItemModel.ListPublishedLearningItemsByNode(courseID, nodeID)
	if err != nil {
		return ctrl.mapLearnerLearningItemError(c, err, constants.ErrListLearningItems)
	}
	response := make([]structs.LearnerLearningItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toLearnerLearningItemResponse(item))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *LearnerLearningItemController) GetByID(c *fiber.Ctx) error {
	user, err := requireAuthenticatedLearner(c)
	if err != nil {
		return mapLearnerAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, itemID, err := parseCourseNodeAndItemIDParams(c)
	if err != nil {
		return err
	}

	if err := ctrl.requireCourseEnrollment(user, courseID); err != nil {
		return ctrl.mapLearnerLearningItemError(c, err, constants.ErrGetLearningItem)
	}

	item, err := ctrl.learningItemModel.GetPublishedLearningItemByID(courseID, nodeID, itemID)
	if err != nil {
		return ctrl.mapLearnerLearningItemError(c, err, constants.ErrGetLearningItem)
	}

	adjacent, err := ctrl.learningItemModel.GetAdjacentPublishedLearningItems(courseID, nodeID, itemID)
	if err != nil {
		return ctrl.mapLearnerLearningItemError(c, err, constants.ErrGetLearningItem)
	}

	var previousLink *structs.LearnerLearningItemNavigation
	if adjacent.Previous != nil {
		previousLink = &structs.LearnerLearningItemNavigation{
			ID:    adjacent.Previous.ID.String(),
			Title: adjacent.Previous.Title,
		}
	}

	var nextLink *structs.LearnerLearningItemNavigation
	if adjacent.Next != nil {
		nextLink = &structs.LearnerLearningItemNavigation{
			ID:    adjacent.Next.ID.String(),
			Title: adjacent.Next.Title,
		}
	}

	response := structs.LearnerLearningItemDetailResponse{
		LearningItem: toLearnerLearningItemResponse(item),
		Previous:     previousLink,
		Next:         nextLink,
	}

	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *LearnerLearningItemController) requireCourseEnrollment(user models.User, courseID uuid.UUID) error {
	return ctrl.enrollmentModel.RequireUserEnrolled(user.ID, courseID)
}

func requireAuthenticatedLearner(c *fiber.Ctx) (models.User, error) {
	user, ok := quizUtilsHelper.ConvertType[models.User](c.Locals(constants.ContextUser))
	if !ok || strings.TrimSpace(user.Email) == "" {
		return models.User{}, errCourseIdentityMissing
	}
	return user, nil
}

func mapLearnerAuthError(c *fiber.Ctx, logger *zap.Logger, err error) error {
	switch {
	case errors.Is(err, errCourseIdentityMissing):
		logger.Error("authenticated learner identity missing")
		return utils.JSONError(c, http.StatusUnauthorized, constants.ErrUnauthenticated)
	default:
		logger.Error("unexpected learner auth error", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUnauthenticated)
	}
}

func (ctrl *LearnerLearningItemController) mapLearnerLearningItemError(c *fiber.Ctx, err error, internalMessage string) error {
	switch {
	case errors.Is(err, models.ErrCourseEnrollmentRequired):
		// Documented denial: no LearningItem payload for unenrolled callers.
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseEnrollmentRequired)
	case errors.Is(err, models.ErrLearningItemNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrLearningItemNotFound)
	case errors.Is(err, models.ErrLearningItemNodeNotFound),
		errors.Is(err, models.ErrCourseNodeNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrLearningItemCrossCourse),
		errors.Is(err, models.ErrCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	default:
		ctrl.logger.Error("learner learning item request failed", zap.Error(err), zap.String("message", internalMessage))
		return utils.JSONError(c, http.StatusInternalServerError, internalMessage)
	}
}

func toLearnerLearningItemResponse(item models.LearningItem) structs.LearnerLearningItemResponse {
	var description *string
	if item.Description.Valid {
		value := item.Description.String
		description = &value
	}
	metadata := json.RawMessage(`null`)
	if len(item.Metadata) > 0 {
		metadata = append(json.RawMessage(nil), item.Metadata...)
	}
	return structs.LearnerLearningItemResponse{
		ID:           item.ID.String(),
		Title:        item.Title,
		ItemType:     string(item.ItemType),
		Description:  description,
		Metadata:     metadata,
		PublishState: string(item.PublishState),
	}
}
