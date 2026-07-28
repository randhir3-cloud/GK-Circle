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

type QuestionCollectionController struct {
	collectionModel *models.QuestionCollectionModel
	appConfig       *config.AppConfig
	logger          *zap.Logger
}

func InitQuestionCollectionController(db *goqu.Database, logger *zap.Logger, appConfig *config.AppConfig) (*QuestionCollectionController, error) {
	return &QuestionCollectionController{
		collectionModel: models.InitQuestionCollectionModel(db),
		appConfig:       appConfig,
		logger:          logger,
	}, nil
}

func (ctrl *QuestionCollectionController) CreateCollection(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}

	var req structs.ReqCreateQuestionCollection
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	kind, err := parseQuestionCollectionKind(req.Kind)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	filterJSON, err := encodeCollectionFilter(kind, req.Filter)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	createdBy, _ := c.Locals(constants.ContextUid).(string)
	collection, err := ctrl.collectionModel.CreateCollection(models.CreateQuestionCollectionParams{
		QuizID:     quizID,
		Title:      req.Title,
		Kind:       kind,
		Position:   req.Position,
		FilterJSON: filterJSON,
		CreatedBy:  createdBy,
	})
	if err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusCreated, toQuestionCollectionResponse(collection))
}

func (ctrl *QuestionCollectionController) ListCollections(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}

	collections, err := ctrl.collectionModel.ListCollectionsByQuizID(quizID)
	if err != nil {
		ctrl.logger.Error("error while listing question collections", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListQuestionCollections)
	}

	response := make([]structs.QuestionCollectionResponse, 0, len(collections))
	for _, collection := range collections {
		response = append(response, toQuestionCollectionResponse(collection))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *QuestionCollectionController) GetCollection(c *fiber.Ctx) error {
	quizID, collectionID, err := parseQuizAndCollectionIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	collection, err := ctrl.collectionModel.GetCollectionWithMembers(quizID, collectionID)
	if err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toQuestionCollectionResponse(collection))
}

func (ctrl *QuestionCollectionController) UpdateCollection(c *fiber.Ctx) error {
	quizID, collectionID, err := parseQuizAndCollectionIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	var req structs.ReqUpdateQuestionCollection
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	params := models.UpdateQuestionCollectionParams{
		Title:    req.Title,
		Position: req.Position,
	}
	if req.Filter != nil {
		filterJSON, err := marshalCollectionFilter(req.Filter)
		if err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuestionCollectionFilterInvalid)
		}
		params.FilterJSON = filterJSON
		params.HasFilter = true
	}

	collection, err := ctrl.collectionModel.UpdateCollection(quizID, collectionID, params)
	if err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toQuestionCollectionResponse(collection))
}

func (ctrl *QuestionCollectionController) DeleteCollection(c *fiber.Ctx) error {
	quizID, collectionID, err := parseQuizAndCollectionIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	if err := ctrl.collectionModel.DeleteCollection(quizID, collectionID); err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, fiber.Map{"deleted": true})
}

func (ctrl *QuestionCollectionController) ReplaceMembers(c *fiber.Ctx) error {
	quizID, collectionID, err := parseQuizAndCollectionIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	var req structs.ReqReplaceQuestionCollectionMembers
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	questionIDs := make([]uuid.UUID, 0, len(req.QuestionIDs))
	for _, rawID := range req.QuestionIDs {
		questionID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuestionInvalidID)
		}
		questionIDs = append(questionIDs, questionID)
	}

	members, err := ctrl.collectionModel.ReplaceStaticMembers(quizID, collectionID, questionIDs)
	if err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}

	collection, err := ctrl.collectionModel.GetCollectionByID(quizID, collectionID)
	if err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}
	collection.Members = members
	return utils.JSONSuccess(c, http.StatusOK, toQuestionCollectionResponse(collection))
}

func (ctrl *QuestionCollectionController) ResolveCollection(c *fiber.Ctx) error {
	quizID, collectionID, err := parseQuizAndCollectionIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	resolution, err := ctrl.collectionModel.ResolveCollection(quizID, collectionID)
	if err != nil {
		return mapQuestionCollectionError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toCollectionResolutionResponse(resolution))
}

func parseQuizIDParam(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(c.Params(constants.QuizId)))
}

func parseQuizAndCollectionIDs(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New(constants.ErrQuizInvalidID)
	}
	collectionID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.CollectionId)))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New(constants.ErrQuestionCollectionInvalidID)
	}
	return quizID, collectionID, nil
}

func encodeCollectionFilter(kind models.QuestionCollectionKind, filter *structs.CollectionDynamicFilter) (json.RawMessage, error) {
	if kind == models.QuestionCollectionKindStatic {
		if filter != nil {
			return nil, errors.New(constants.ErrQuestionCollectionFilterInvalid)
		}
		return nil, nil
	}
	if filter == nil {
		filter = &structs.CollectionDynamicFilter{}
	}
	return marshalCollectionFilter(filter)
}

func marshalCollectionFilter(filter *structs.CollectionDynamicFilter) (json.RawMessage, error) {
	modelFilter := models.CollectionDynamicFilter{
		Subject:    filter.Subject,
		Topic:      filter.Topic,
		Year:       filter.Year,
		Difficulty: filter.Difficulty,
		PYQStatus:  filter.PYQStatus,
	}
	return json.Marshal(modelFilter)
}

func parseQuestionCollectionKind(raw string) (models.QuestionCollectionKind, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(models.QuestionCollectionKindStatic):
		return models.QuestionCollectionKindStatic, nil
	case string(models.QuestionCollectionKindDynamic):
		return models.QuestionCollectionKindDynamic, nil
	default:
		return "", models.ErrQuestionCollectionKindInvalid
	}
}

func toStructCollectionFilter(filter *models.CollectionDynamicFilter) *structs.CollectionDynamicFilter {
	if filter == nil {
		return nil
	}
	return &structs.CollectionDynamicFilter{
		Subject:    filter.Subject,
		Topic:      filter.Topic,
		Year:       filter.Year,
		Difficulty: filter.Difficulty,
		PYQStatus:  filter.PYQStatus,
	}
}

func toQuestionCollectionResponse(collection models.QuestionCollection) structs.QuestionCollectionResponse {
	response := structs.QuestionCollectionResponse{
		ID:        collection.ID.String(),
		QuizID:    collection.QuizID.String(),
		Title:     collection.Title,
		Kind:      string(collection.Kind),
		Position:  collection.Position,
		Filter:    toStructCollectionFilter(collection.Filter),
		CreatedBy: collection.CreatedBy,
		CreatedAt: collection.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: collection.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if len(collection.Members) > 0 {
		response.Members = make([]structs.QuestionCollectionMemberResponse, 0, len(collection.Members))
		for _, member := range collection.Members {
			response.Members = append(response.Members, structs.QuestionCollectionMemberResponse{
				ID:         member.ID.String(),
				QuestionID: member.QuestionID.String(),
				Position:   member.Position,
			})
		}
	}
	return response
}

func toCollectionResolutionResponse(resolution models.CollectionResolution) structs.CollectionResolutionResponse {
	questionIDs := make([]string, 0, len(resolution.QuestionIDs))
	for _, questionID := range resolution.QuestionIDs {
		questionIDs = append(questionIDs, questionID.String())
	}
	return structs.CollectionResolutionResponse{
		CollectionID:     resolution.CollectionID.String(),
		Kind:             string(resolution.Kind),
		QuestionIDs:      questionIDs,
		ResolutionStatus: resolution.ResolutionStatus,
		Message:          resolution.Message,
	}
}

func mapQuestionCollectionError(c *fiber.Ctx, logger *zap.Logger, err error) error {
	switch {
	case errors.Is(err, models.ErrQuestionCollectionNotFound):
		return utils.JSONFail(c, http.StatusNotFound, constants.ErrQuestionCollectionNotFound)
	case errors.Is(err, models.ErrQuestionCollectionTitleRequired),
		errors.Is(err, models.ErrQuestionCollectionKindInvalid),
		errors.Is(err, models.ErrQuestionCollectionFilterRequired),
		errors.Is(err, models.ErrQuestionCollectionFilterInvalid),
		errors.Is(err, models.ErrQuestionCollectionKindMismatch),
		errors.Is(err, models.ErrQuestionCollectionMemberInvalid),
		errors.Is(err, models.ErrQuestionCollectionMemberDuplicate),
		errors.Is(err, models.ErrQuestionCollectionUpdateRequired):
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	default:
		logger.Error("question collection operation failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetQuestionCollection)
	}
}
