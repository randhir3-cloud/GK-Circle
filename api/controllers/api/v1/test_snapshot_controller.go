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
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

type TestSnapshotController struct {
	snapshotSvc *services.TestSnapshotService
	appConfig   *config.AppConfig
	logger      *zap.Logger
}

func InitTestSnapshotController(db *goqu.Database, logger *zap.Logger, appConfig *config.AppConfig) (*TestSnapshotController, error) {
	return &TestSnapshotController{
		snapshotSvc: services.NewTestSnapshotService(db, logger),
		appConfig:   appConfig,
		logger:      logger,
	}, nil
}

func (ctrl *TestSnapshotController) CreateSnapshot(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}

	var req structs.ReqCreateTestSnapshot
	if len(c.Body()) > 0 {
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, err.Error())
		}
	}

	collectionIDs := make([]uuid.UUID, 0, len(req.CollectionIDs))
	for _, raw := range req.CollectionIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuestionCollectionInvalidID)
		}
		collectionIDs = append(collectionIDs, id)
	}

	createdBy, _ := c.Locals(constants.ContextUid).(string)
	snapshot, err := ctrl.snapshotSvc.CreateFromCollections(services.CreateTestSnapshotRequest{
		QuizID:        quizID,
		CreatedBy:     createdBy,
		CollectionIDs: collectionIDs,
	})
	if err != nil {
		return mapTestSnapshotError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusCreated, toTestSnapshotResponse(snapshot))
}

func (ctrl *TestSnapshotController) ListSnapshots(c *fiber.Ctx) error {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, constants.ErrQuizInvalidID)
	}

	snapshots, err := ctrl.snapshotSvc.List(quizID)
	if err != nil {
		ctrl.logger.Error("error while listing test snapshots", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrListTestSnapshots)
	}
	response := make([]structs.TestSnapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		response = append(response, toTestSnapshotResponse(snapshot))
	}
	return utils.JSONSuccess(c, http.StatusOK, response)
}

func (ctrl *TestSnapshotController) GetSnapshot(c *fiber.Ctx) error {
	quizID, snapshotID, err := parseQuizAndSnapshotIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	snapshot, err := ctrl.snapshotSvc.Get(quizID, snapshotID)
	if err != nil {
		return mapTestSnapshotError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toTestSnapshotResponse(snapshot))
}

func (ctrl *TestSnapshotController) GetLearnerSnapshot(c *fiber.Ctx) error {
	quizID, snapshotID, err := parseQuizAndSnapshotIDs(c)
	if err != nil {
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	}

	view, err := ctrl.snapshotSvc.GetLearnerView(quizID, snapshotID)
	if err != nil {
		return mapTestSnapshotError(c, ctrl.logger, err)
	}
	return utils.JSONSuccess(c, http.StatusOK, toTestSnapshotLearnerResponse(view))
}

func parseQuizAndSnapshotIDs(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	quizID, err := parseQuizIDParam(c)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New(constants.ErrQuizInvalidID)
	}
	snapshotID, err := uuid.Parse(strings.TrimSpace(c.Params(constants.SnapshotId)))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New(constants.ErrTestSnapshotInvalidID)
	}
	return quizID, snapshotID, nil
}

func toTestSnapshotResponse(snapshot models.TestSnapshot) structs.TestSnapshotResponse {
	sourceIDs := make([]string, 0, len(snapshot.SourceCollectionIDs))
	for _, id := range snapshot.SourceCollectionIDs {
		sourceIDs = append(sourceIDs, id.String())
	}
	response := structs.TestSnapshotResponse{
		ID:                  snapshot.ID.String(),
		QuizID:              snapshot.QuizID.String(),
		CreatedBy:           snapshot.CreatedBy,
		Status:              snapshot.Status,
		SourceCollectionIDs: sourceIDs,
		QuestionCount:       snapshot.QuestionCount,
		CreatedAt:           snapshot.CreatedAt.UTC().Format(time.RFC3339),
	}
	if len(snapshot.Items) > 0 {
		response.Items = make([]structs.TestSnapshotItemResponse, 0, len(snapshot.Items))
		for _, item := range snapshot.Items {
			entry := structs.TestSnapshotItemResponse{
				ID:                  item.ID.String(),
				Position:            item.Position,
				QuestionID:          item.QuestionID.String(),
				LineageID:           item.LineageID.String(),
				RevisionNumber:      item.RevisionNumber,
				Question:            item.Question,
				Type:                item.Type,
				Options:             item.Options,
				Answers:             item.Answers,
				OfficialAnswer:      item.OfficialAnswer,
				AuthoritativeAnswer: item.AuthoritativeAnswer,
				AnswerReviewStatus:  item.AnswerReviewStatus,
				Points:              item.Points,
				DurationInSeconds:   item.DurationInSeconds,
				QuestionMedia:       item.QuestionMedia,
				OptionsMedia:        item.OptionsMedia,
				Resource:            item.Resource,
			}
			if item.CollectionID != nil {
				value := item.CollectionID.String()
				entry.CollectionID = &value
			}
			response.Items = append(response.Items, entry)
		}
	}
	return response
}

func toTestSnapshotLearnerResponse(view models.TestSnapshotLearnerView) structs.TestSnapshotLearnerResponse {
	sourceIDs := make([]string, 0, len(view.SourceCollectionIDs))
	for _, id := range view.SourceCollectionIDs {
		sourceIDs = append(sourceIDs, id.String())
	}
	items := make([]structs.TestSnapshotLearnerItemResponse, 0, len(view.Items))
	for _, item := range view.Items {
		items = append(items, structs.TestSnapshotLearnerItemResponse{
			Position:          item.Position,
			QuestionID:        item.QuestionID.String(),
			LineageID:         item.LineageID.String(),
			RevisionNumber:    item.RevisionNumber,
			Question:          item.Question,
			Type:              item.Type,
			Options:           item.Options,
			Points:            item.Points,
			DurationInSeconds: item.DurationInSeconds,
			QuestionMedia:     item.QuestionMedia,
			OptionsMedia:      item.OptionsMedia,
			Resource:          item.Resource,
		})
	}
	return structs.TestSnapshotLearnerResponse{
		ID:                  view.ID.String(),
		QuizID:              view.QuizID.String(),
		Status:              view.Status,
		SourceCollectionIDs: sourceIDs,
		QuestionCount:       view.QuestionCount,
		CreatedAt:           view.CreatedAt.UTC().Format(time.RFC3339),
		Items:               items,
	}
}

func mapTestSnapshotError(c *fiber.Ctx, logger *zap.Logger, err error) error {
	switch {
	case errors.Is(err, models.ErrTestSnapshotNotFound),
		errors.Is(err, models.ErrQuestionCollectionNotFound):
		return utils.JSONFail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrTestSnapshotEmpty),
		errors.Is(err, models.ErrTestSnapshotNoCollections),
		errors.Is(err, models.ErrTestSnapshotUnresolvedCollection),
		errors.Is(err, models.ErrTestSnapshotDuplicateQuestion),
		errors.Is(err, models.ErrTestSnapshotQuestionMissing):
		return utils.JSONFail(c, http.StatusBadRequest, err.Error())
	default:
		logger.Error("test snapshot operation failed", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetTestSnapshot)
	}
}
