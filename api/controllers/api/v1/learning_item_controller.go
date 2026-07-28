package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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

type LearningItemController struct {
	learningItemModel *models.LearningItemModel
	appConfig         *config.AppConfig
	logger            *zap.Logger
}

func InitLearningItemController(db *goqu.Database, logger *zap.Logger, appConfig *config.AppConfig) (*LearningItemController, error) {
	return &LearningItemController{
		learningItemModel: models.InitLearningItemModel(db),
		appConfig:         appConfig,
		logger:            logger,
	}, nil
}

func (ctrl *LearningItemController) Create(c *fiber.Ctx) error {
	_, actorID, err := authorizeCourseAdmin(c, ctrl.appConfig)
	if err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, err := parseCourseAndNodeIDParams(c)
	if err != nil {
		return err
	}

	var req structs.ReqCreateAdminLearningItem
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	params, err := validateCreateAdminLearningItem(courseID, nodeID, actorID, req)
	if err != nil {
		if errors.Is(err, models.ErrLearningItemPublishStateInvalid) {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemPublishState)
		}
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	item, err := ctrl.learningItemModel.CreateLearningItem(params)
	if err != nil {
		return ctrl.mapLearningItemError(c, err, constants.ErrCreateLearningItem)
	}
	return utils.JSONSuccess(c, http.StatusCreated, toAdminLearningItemResponse(item))
}

func (ctrl *LearningItemController) List(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, err := parseCourseAndNodeIDParams(c)
	if err != nil {
		return err
	}

	items, err := ctrl.learningItemModel.ListLearningItemsByNode(courseID, nodeID)
	if err != nil {
		return ctrl.mapLearningItemError(c, err, constants.ErrListLearningItems)
	}
	response := make([]structs.AdminLearningItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toAdminLearningItemResponse(item))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *LearningItemController) Reorder(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, err := parseCourseAndNodeIDParams(c)
	if err != nil {
		return err
	}

	var req structs.ReqReorderAdminLearningItems
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	orderedIDs, err := parseOrderedLearningItemIDs(req.OrderedItemIDs)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	result, err := ctrl.learningItemModel.ReorderLearningItems(courseID, nodeID, orderedIDs)
	if err != nil {
		return ctrl.mapLearningItemReorderError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, structs.ResReorderAdminLearningItems{
		CourseNodeID:      result.CourseNodeID.String(),
		LearningItemCount: result.LearningItemCount,
		PositionsUpdated:  result.PositionsUpdated,
		Noop:              result.Noop,
	})
}

func (ctrl *LearningItemController) Move(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, sourceNodeID, err := parseCourseAndNodeIDParams(c)
	if err != nil {
		return err
	}

	var req structs.ReqMoveAdminLearningItems
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	destinationNodeID, err := uuid.Parse(strings.TrimSpace(req.DestinationNodeID))
	if err != nil || destinationNodeID == uuid.Nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemMoveInvalid)
	}

	orderedIDs, err := parseOrderedLearningItemIDs(req.OrderedItemIDs)
	if err != nil {
		if err.Error() == constants.ErrLearningItemReorderInvalid {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemMoveInvalid)
		}
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemMoveMismatch)
	}

	result, err := ctrl.learningItemModel.MoveLearningItems(
		courseID, sourceNodeID, destinationNodeID, orderedIDs,
	)
	if err != nil {
		return ctrl.mapLearningItemMoveError(c, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, structs.ResMoveLearningItems{
		SourceNodeID:         result.SourceNodeID.String(),
		DestinationNodeID:    result.DestinationNodeID.String(),
		ItemsMoved:           result.ItemsMoved,
		SourceItemCount:      result.SourceItemCount,
		DestinationItemCount: result.DestinationItemCount,
		Noop:                 result.Noop,
	})
}

func (ctrl *LearningItemController) GetByID(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, itemID, err := parseCourseNodeAndItemIDParams(c)
	if err != nil {
		return err
	}

	item, err := ctrl.learningItemModel.GetLearningItemByID(courseID, nodeID, itemID)
	if err != nil {
		return ctrl.mapLearningItemError(c, err, constants.ErrGetLearningItem)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAdminLearningItemResponse(item))
}

func (ctrl *LearningItemController) Update(c *fiber.Ctx) error {
	_, actorID, err := authorizeCourseAdmin(c, ctrl.appConfig)
	if err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, itemID, err := parseCourseNodeAndItemIDParams(c)
	if err != nil {
		return err
	}

	var req structs.ReqUpdateAdminLearningItem
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	params, err := validateUpdateAdminLearningItem(actorID, req)
	if err != nil {
		if errors.Is(err, models.ErrLearningItemUpdateRequired) {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemEmptyPatch)
		}
		if errors.Is(err, models.ErrLearningItemPublishStateInvalid) {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemPublishState)
		}
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	item, err := ctrl.learningItemModel.UpdateLearningItem(courseID, nodeID, itemID, params)
	if err != nil {
		return ctrl.mapLearningItemError(c, err, constants.ErrUpdateLearningItem)
	}
	return utils.JSONSuccess(c, http.StatusOK, toAdminLearningItemResponse(item))
}

func (ctrl *LearningItemController) Delete(c *fiber.Ctx) error {
	if _, _, err := authorizeCourseAdmin(c, ctrl.appConfig); err != nil {
		return mapCourseAuthError(c, ctrl.logger, err)
	}

	courseID, nodeID, itemID, err := parseCourseNodeAndItemIDParams(c)
	if err != nil {
		return err
	}

	if err := ctrl.learningItemModel.DeleteLearningItem(courseID, nodeID, itemID); err != nil {
		return ctrl.mapLearningItemError(c, err, constants.ErrDeleteLearningItem)
	}
	return utils.JSONSuccess(c, http.StatusOK, "success")
}

func (ctrl *LearningItemController) mapLearningItemError(c *fiber.Ctx, err error, internalMessage string) error {
	switch {
	case errors.Is(err, models.ErrLearningItemQuizRequired),
		errors.Is(err, models.ErrLearningItemQuizForbidden):
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrLearningItemQuizNotFound):
		return utils.JSONFail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrLearningItemNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrLearningItemNotFound)
	case errors.Is(err, models.ErrLearningItemNodeNotFound),
		errors.Is(err, models.ErrCourseNodeNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrLearningItemCrossCourse),
		errors.Is(err, models.ErrCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrLearningItemUpdateRequired):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemEmptyPatch)
	case errors.Is(err, models.ErrLearningItemTitleRequired):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemTitleInvalid)
	case errors.Is(err, models.ErrLearningItemTypeInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemTypeInvalid)
	case errors.Is(err, models.ErrLearningItemMetadataInvalid),
		errors.Is(err, models.ErrLearningItemBlockDuplicate),
		errors.Is(err, models.ErrLearningItemBlockTypeInvalid),
		errors.Is(err, models.ErrLearningItemMetadataVersionInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemMetadataInvalid)
	case errors.Is(err, models.ErrLearningItemPlaceholderSyntax),
		errors.Is(err, models.ErrLearningItemPlaceholderInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemPlaceholder)
	case errors.Is(err, models.ErrLearningItemVisibilityInvalid),
		errors.Is(err, models.ErrLearningItemVisibilityModeInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemVisibility)
	case errors.Is(err, models.ErrLearningItemPublishStateInvalid):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemPublishState)
	case errors.Is(err, models.ErrLearningItemConflict),
		errors.Is(err, models.ErrLearningItemPositionInvalid):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrLearningItemConflict)
	default:
		ctrl.logger.Error("learning item request failed", zap.Error(err), zap.String("message", internalMessage))
		return utils.JSONError(c, http.StatusInternalServerError, internalMessage)
	}
}

func (ctrl *LearningItemController) mapLearningItemReorderError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrLearningItemReorderDuplicate),
		errors.Is(err, models.ErrLearningItemReorderMismatch),
		errors.Is(err, models.ErrLearningItemNotFound):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemReorderMismatch)
	case errors.Is(err, models.ErrLearningItemNodeNotFound),
		errors.Is(err, models.ErrCourseNodeNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrLearningItemCrossCourse),
		errors.Is(err, models.ErrCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrLearningItemReorderConflict),
		errors.Is(err, models.ErrLearningItemPositionInvalid):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrLearningItemReorderConflict)
	default:
		ctrl.logger.Error("error while reordering learning items", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrReorderLearningItems)
	}
}

func (ctrl *LearningItemController) mapLearningItemMoveError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, models.ErrLearningItemMoveDuplicate),
		errors.Is(err, models.ErrLearningItemMoveMismatch),
		errors.Is(err, models.ErrLearningItemNotFound):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemMoveMismatch)
	case errors.Is(err, models.ErrLearningItemMoveSameNode):
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemMoveSameNode)
	case errors.Is(err, models.ErrLearningItemNodeNotFound),
		errors.Is(err, models.ErrCourseNodeNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNodeNotFound)
	case errors.Is(err, models.ErrLearningItemCrossCourse),
		errors.Is(err, models.ErrCourseNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrCourseNotFound)
	case errors.Is(err, models.ErrLearningItemMoveConflict),
		errors.Is(err, models.ErrLearningItemPositionInvalid):
		return utils.JSONFail(c, http.StatusConflict, constants.ErrLearningItemMoveConflict)
	default:
		ctrl.logger.Error("error while moving learning items", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrMoveLearningItems)
	}
}

func parseOrderedLearningItemIDs(rawIDs []string) ([]uuid.UUID, error) {
	if rawIDs == nil {
		return []uuid.UUID{}, nil
	}
	ordered := make([]uuid.UUID, 0, len(rawIDs))
	seen := make(map[uuid.UUID]struct{}, len(rawIDs))
	for _, raw := range rawIDs {
		itemID, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil || itemID == uuid.Nil {
			return nil, errors.New(constants.ErrLearningItemReorderInvalid)
		}
		if _, exists := seen[itemID]; exists {
			return nil, errors.New(constants.ErrLearningItemReorderMismatch)
		}
		seen[itemID] = struct{}{}
		ordered = append(ordered, itemID)
	}
	return ordered, nil
}

func parseCourseAndNodeIDParams(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	courseID, err := parseCourseIDParam(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseInvalidID)
	}
	nodeID, err := parseNodeIDParam(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, utils.JSONFail(c, http.StatusBadRequest, constants.ErrCourseNodeInvalidID)
	}
	return courseID, nodeID, nil
}

func parseCourseNodeAndItemIDParams(c *fiber.Ctx) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	courseID, nodeID, err := parseCourseAndNodeIDParams(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	itemID, err := parseItemIDParam(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, utils.JSONFail(c, http.StatusBadRequest, constants.ErrLearningItemInvalidID)
	}
	return courseID, nodeID, itemID, nil
}

func parseItemIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(c.Params(constants.ItemId)))
}

func validateCreateAdminLearningItem(
	courseID, nodeID uuid.UUID,
	actorID string,
	req structs.ReqCreateAdminLearningItem,
) (models.CreateLearningItemParams, error) {
	params := models.CreateLearningItemParams{
		CourseID:     courseID,
		CourseNodeID: nodeID,
		ActorID:      actorID,
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return params, errors.New(constants.ErrLearningItemTitleInvalid)
	}
	params.Title = title

	itemType := strings.TrimSpace(req.ItemType)
	if itemType == "" {
		return params, errors.New(constants.ErrLearningItemTypeInvalid)
	}
	params.ItemType = models.LearningItemType(itemType)

	if req.QuizID.Present && !req.QuizID.Null {
		quizID, err := uuid.Parse(strings.TrimSpace(req.QuizID.Value))
		if err != nil || quizID == uuid.Nil {
			return params, models.ErrLearningItemQuizRequired
		}
		params.QuizID = &quizID
	}
	if params.ItemType == models.LearningItemTypeQuizRef && params.QuizID == nil {
		return params, models.ErrLearningItemQuizRequired
	}
	if params.ItemType != models.LearningItemTypeQuizRef && params.QuizID != nil {
		return params, models.ErrLearningItemQuizForbidden
	}

	if req.Description.Present && !req.Description.Null {
		params.Description = strings.TrimSpace(req.Description.Value)
	}

	if req.Metadata.Present && !req.Metadata.Null {
		params.Metadata = append(json.RawMessage(nil), req.Metadata.Value...)
	}

	if req.PublishState.Present {
		if req.PublishState.Null {
			return params, models.ErrLearningItemPublishStateInvalid
		}
		state := models.LearningItemPublishState(strings.TrimSpace(req.PublishState.Value))
		if state != models.LearningItemPublishStateDraft &&
			state != models.LearningItemPublishStatePublished {
			return params, models.ErrLearningItemPublishStateInvalid
		}
		params.PublishState = state
	}

	return params, nil
}

func validateUpdateAdminLearningItem(actorID string, req structs.ReqUpdateAdminLearningItem) (models.UpdateLearningItemParams, error) {
	params := models.UpdateLearningItemParams{ActorID: actorID}
	supplied := false

	if req.Title.Present {
		supplied = true
		if req.Title.Null {
			return params, errors.New(constants.ErrLearningItemTitleInvalid)
		}
		title := strings.TrimSpace(req.Title.Value)
		if title == "" {
			return params, errors.New(constants.ErrLearningItemTitleInvalid)
		}
		params.Title = &title
	}

	if req.ItemType.Present {
		supplied = true
		if req.ItemType.Null {
			return params, errors.New(constants.ErrLearningItemTypeInvalid)
		}
		itemType := models.LearningItemType(strings.TrimSpace(req.ItemType.Value))
		if itemType == "" {
			return params, errors.New(constants.ErrLearningItemTypeInvalid)
		}
		params.ItemType = &itemType
	}

	if req.Description.Present {
		supplied = true
		params.Description = models.OptionalNullableString{
			Present: true,
			Null:    req.Description.Null,
			Value:   strings.TrimSpace(req.Description.Value),
		}
		if req.Description.Null {
			params.Description.Value = ""
		}
	}

	if req.Metadata.Present {
		supplied = true
		params.Metadata = models.OptionalJSONBytes{
			Present: true,
			Null:    req.Metadata.Null,
			Value:   append(json.RawMessage(nil), req.Metadata.Value...),
		}
	}

	if req.QuizID.Present {
		supplied = true
		params.QuizID.Present = true
		params.QuizID.Null = req.QuizID.Null
		if !req.QuizID.Null {
			quizID, err := uuid.Parse(strings.TrimSpace(req.QuizID.Value))
			if err != nil || quizID == uuid.Nil {
				return params, models.ErrLearningItemQuizRequired
			}
			params.QuizID.Value = quizID
		}
	}

	if params.ItemType != nil {
		if *params.ItemType == models.LearningItemTypeQuizRef &&
			(!params.QuizID.Present || params.QuizID.Null) {
			return params, models.ErrLearningItemQuizRequired
		}
		if *params.ItemType != models.LearningItemTypeQuizRef &&
			params.QuizID.Present && !params.QuizID.Null {
			return params, models.ErrLearningItemQuizForbidden
		}
	}

	if req.PublishState.Present {
		supplied = true
		if req.PublishState.Null {
			return params, models.ErrLearningItemPublishStateInvalid
		}
		state := models.LearningItemPublishState(strings.TrimSpace(req.PublishState.Value))
		if state != models.LearningItemPublishStateDraft &&
			state != models.LearningItemPublishStatePublished {
			return params, models.ErrLearningItemPublishStateInvalid
		}
		params.PublishState = &state
	}

	if !supplied {
		return params, models.ErrLearningItemUpdateRequired
	}
	return params, nil
}

func toAdminLearningItemResponse(item models.LearningItem) structs.AdminLearningItemResponse {
	var description *string
	if item.Description.Valid {
		value := item.Description.String
		description = &value
	}
	metadata := json.RawMessage(`null`)
	if len(item.Metadata) > 0 {
		metadata = append(json.RawMessage(nil), item.Metadata...)
	}
	return structs.AdminLearningItemResponse{
		ID:           item.ID.String(),
		CourseID:     item.CourseID.String(),
		CourseNodeID: item.CourseNodeID.String(),
		Title:        item.Title,
		ItemType:     string(item.ItemType),
		Description:  description,
		Metadata:     metadata,
		Position:     item.Position,
		PublishState: string(item.PublishState),
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
