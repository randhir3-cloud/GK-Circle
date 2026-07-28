package services

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

// TestSnapshotService composes immutable quiz snapshots from collections (EXAM-P4-T04).
// It does not start attempts or score answers (EXAM-P5).
type TestSnapshotService struct {
	collectionModel *models.QuestionCollectionModel
	snapshotModel   *models.TestSnapshotModel
	logger          *zap.Logger
}

func NewTestSnapshotService(db *goqu.Database, logger *zap.Logger) *TestSnapshotService {
	return &TestSnapshotService{
		collectionModel: models.InitQuestionCollectionModel(db),
		snapshotModel:   models.InitTestSnapshotModel(db),
		logger:          logger,
	}
}

type CreateTestSnapshotRequest struct {
	QuizID        uuid.UUID
	CreatedBy     string
	CollectionIDs []uuid.UUID // optional; empty means all quiz collections in position order
}

func (svc *TestSnapshotService) CreateFromCollections(req CreateTestSnapshotRequest) (models.TestSnapshot, error) {
	collections, err := svc.collectionModel.ListCollectionsByQuizID(req.QuizID)
	if err != nil {
		return models.TestSnapshot{}, err
	}
	if len(collections) == 0 {
		return models.TestSnapshot{}, models.ErrTestSnapshotNoCollections
	}

	selected := collections
	if len(req.CollectionIDs) > 0 {
		wanted := make(map[uuid.UUID]struct{}, len(req.CollectionIDs))
		for _, id := range req.CollectionIDs {
			wanted[id] = struct{}{}
		}
		selected = make([]models.QuestionCollection, 0, len(req.CollectionIDs))
		for _, collection := range collections {
			if _, ok := wanted[collection.ID]; ok {
				selected = append(selected, collection)
				delete(wanted, collection.ID)
			}
		}
		if len(wanted) > 0 {
			return models.TestSnapshot{}, models.ErrQuestionCollectionNotFound
		}
	}

	sourceIDs := make([]uuid.UUID, 0, len(selected))
	composed := make([]models.CreateTestSnapshotItemParams, 0)
	seenQuestions := make(map[uuid.UUID]struct{})
	allQuestionIDs := make([]uuid.UUID, 0)

	for _, collection := range selected {
		sourceIDs = append(sourceIDs, collection.ID)
		resolution, err := svc.collectionModel.ResolveCollection(req.QuizID, collection.ID)
		if err != nil {
			return models.TestSnapshot{}, err
		}
		if resolution.ResolutionStatus == models.CollectionResolutionStatusMetadataPending {
			return models.TestSnapshot{}, models.ErrTestSnapshotUnresolvedCollection
		}
		if len(resolution.QuestionIDs) == 0 {
			continue
		}
		collectionID := collection.ID
		for _, questionID := range resolution.QuestionIDs {
			if _, exists := seenQuestions[questionID]; exists {
				return models.TestSnapshot{}, models.ErrTestSnapshotDuplicateQuestion
			}
			seenQuestions[questionID] = struct{}{}
			allQuestionIDs = append(allQuestionIDs, questionID)
			composed = append(composed, models.CreateTestSnapshotItemParams{
				Position:     len(composed),
				CollectionID: &collectionID,
				Freeze: models.SnapshotQuestionFreeze{
					QuestionID: questionID,
				},
			})
		}
	}

	if len(composed) == 0 {
		return models.TestSnapshot{}, models.ErrTestSnapshotEmpty
	}

	freezes, err := svc.snapshotModel.LoadQuestionFreezesByIDs(req.QuizID, allQuestionIDs)
	if err != nil {
		return models.TestSnapshot{}, err
	}
	for i := range composed {
		freeze, ok := freezes[composed[i].Freeze.QuestionID]
		if !ok {
			return models.TestSnapshot{}, models.ErrTestSnapshotQuestionMissing
		}
		composed[i].Freeze = freeze
	}

	return svc.snapshotModel.CreateSnapshot(models.CreateTestSnapshotParams{
		QuizID:              req.QuizID,
		CreatedBy:           req.CreatedBy,
		SourceCollectionIDs: sourceIDs,
		Items:               composed,
	})
}

func (svc *TestSnapshotService) List(quizID uuid.UUID) ([]models.TestSnapshot, error) {
	return svc.snapshotModel.ListSnapshotsByQuizID(quizID)
}

func (svc *TestSnapshotService) Get(quizID, snapshotID uuid.UUID) (models.TestSnapshot, error) {
	return svc.snapshotModel.GetSnapshotByID(quizID, snapshotID)
}

func (svc *TestSnapshotService) GetLearnerView(quizID, snapshotID uuid.UUID) (models.TestSnapshotLearnerView, error) {
	snapshot, err := svc.snapshotModel.GetSnapshotByID(quizID, snapshotID)
	if err != nil {
		return models.TestSnapshotLearnerView{}, err
	}
	return svc.snapshotModel.ToLearnerView(snapshot), nil
}
