package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

const TestSnapshotsTable = "test_snapshots"
const TestSnapshotItemsTable = "test_snapshot_items"

const TestSnapshotStatusCreated = "CREATED"

var (
	ErrTestSnapshotNotFound            = errors.New("test snapshot not found")
	ErrTestSnapshotEmpty               = errors.New("test snapshot requires at least one resolved question")
	ErrTestSnapshotUnresolvedCollection = errors.New("test snapshot blocked: collection filters are not fully resolved")
	ErrTestSnapshotDuplicateQuestion   = errors.New("test snapshot blocked: duplicate question across collections")
	ErrTestSnapshotNoCollections       = errors.New("test snapshot requires at least one collection")
	ErrTestSnapshotQuestionMissing     = errors.New("test snapshot blocked: resolved question not found in quiz bank")
)

type TestSnapshot struct {
	ID                   uuid.UUID   `json:"id" db:"id"`
	QuizID               uuid.UUID   `json:"quiz_id" db:"quiz_id"`
	CreatedBy            string      `json:"created_by" db:"created_by"`
	Status               string      `json:"status" db:"status"`
	SourceCollectionJSON []byte      `json:"-" db:"source_collection_ids"`
	SourceCollectionIDs  []uuid.UUID `json:"source_collection_ids" db:"-"`
	QuestionCount        int         `json:"question_count" db:"question_count"`
	CreatedAt            time.Time   `json:"created_at" db:"created_at"`
	Items                []TestSnapshotItem `json:"items,omitempty" db:"-"`
}

type TestSnapshotItem struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	SnapshotID           uuid.UUID  `json:"snapshot_id" db:"snapshot_id"`
	Position             int        `json:"position" db:"position"`
	CollectionID         *uuid.UUID `json:"collection_id,omitempty" db:"collection_id"`
	QuestionID           uuid.UUID  `json:"question_id" db:"question_id"`
	LineageID            uuid.UUID  `json:"lineage_id" db:"lineage_id"`
	RevisionNumber       int        `json:"revision_number" db:"revision_number"`
	Question             string     `json:"question" db:"question"`
	Type                 int        `json:"type" db:"type"`
	OptionsJSON          []byte     `json:"-" db:"options"`
	AnswersJSON          []byte     `json:"-" db:"answers"`
	OfficialAnswerJSON   []byte     `json:"-" db:"official_answer"`
	AuthoritativeJSON    []byte     `json:"-" db:"authoritative_answer"`
	Options              map[string]string `json:"options" db:"-"`
	Answers              []int      `json:"answers,omitempty" db:"-"`
	OfficialAnswer       []int      `json:"official_answer,omitempty" db:"-"`
	AuthoritativeAnswer  []int      `json:"authoritative_answer,omitempty" db:"-"`
	AnswerReviewStatus   string     `json:"answer_review_status" db:"answer_review_status"`
	Points               *int16     `json:"points,omitempty" db:"points"`
	DurationInSeconds    *int       `json:"duration_in_seconds,omitempty" db:"duration_in_seconds"`
	QuestionMedia        string     `json:"question_media,omitempty" db:"question_media"`
	OptionsMedia         string     `json:"options_media,omitempty" db:"options_media"`
	Resource             *string    `json:"resource,omitempty" db:"resource"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

// TestSnapshotLearnerItem is answer-key-safe for future player consumption.
type TestSnapshotLearnerItem struct {
	Position          int               `json:"position"`
	QuestionID        uuid.UUID         `json:"question_id"`
	LineageID         uuid.UUID         `json:"lineage_id"`
	RevisionNumber    int               `json:"revision_number"`
	Question          string            `json:"question"`
	Type              int               `json:"type"`
	Options           map[string]string `json:"options"`
	Points            *int16            `json:"points,omitempty"`
	DurationInSeconds *int              `json:"duration_in_seconds,omitempty"`
	QuestionMedia     string            `json:"question_media,omitempty"`
	OptionsMedia      string            `json:"options_media,omitempty"`
	Resource          *string           `json:"resource,omitempty"`
}

type TestSnapshotLearnerView struct {
	ID                  uuid.UUID                `json:"id"`
	QuizID              uuid.UUID                `json:"quiz_id"`
	Status              string                   `json:"status"`
	SourceCollectionIDs []uuid.UUID              `json:"source_collection_ids"`
	QuestionCount       int                      `json:"question_count"`
	CreatedAt           time.Time                `json:"created_at"`
	Items               []TestSnapshotLearnerItem `json:"items"`
}

type SnapshotFreezeSource struct {
	CollectionID uuid.UUID
	QuestionID   uuid.UUID
	Position     int
}

type SnapshotQuestionFreeze struct {
	QuestionID          uuid.UUID
	LineageID           uuid.UUID
	RevisionNumber      int
	Question            string
	Type                int
	OptionsJSON         []byte
	AnswersJSON         []byte
	OfficialAnswerJSON  []byte
	AuthoritativeJSON   []byte
	AnswerReviewStatus  string
	Points              *int16
	DurationInSeconds   *int
	QuestionMedia       string
	OptionsMedia        string
	Resource            *string
}

type CreateTestSnapshotParams struct {
	QuizID              uuid.UUID
	CreatedBy           string
	SourceCollectionIDs []uuid.UUID
	Items               []CreateTestSnapshotItemParams
}

type CreateTestSnapshotItemParams struct {
	Position      int
	CollectionID  *uuid.UUID
	Freeze        SnapshotQuestionFreeze
}

type TestSnapshotModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitTestSnapshotModel(goquDB *goqu.Database) *TestSnapshotModel {
	return &TestSnapshotModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

// SetUUIDGenerator overrides UUID creation for deterministic tests.
func (model *TestSnapshotModel) SetUUIDGenerator(fn func() (uuid.UUID, error)) {
	if fn != nil {
		model.newUUID = fn
	}
}

func (model *TestSnapshotModel) CreateSnapshot(params CreateTestSnapshotParams) (TestSnapshot, error) {
	if len(params.Items) == 0 {
		return TestSnapshot{}, ErrTestSnapshotEmpty
	}
	if len(params.SourceCollectionIDs) == 0 {
		return TestSnapshot{}, ErrTestSnapshotNoCollections
	}

	seen := make(map[uuid.UUID]struct{}, len(params.Items))
	for _, item := range params.Items {
		if _, ok := seen[item.Freeze.QuestionID]; ok {
			return TestSnapshot{}, ErrTestSnapshotDuplicateQuestion
		}
		seen[item.Freeze.QuestionID] = struct{}{}
	}

	snapshotID, err := model.newUUID()
	if err != nil {
		return TestSnapshot{}, err
	}

	sourceJSON, err := json.Marshal(params.SourceCollectionIDs)
	if err != nil {
		return TestSnapshot{}, err
	}

	tx, err := model.db.Begin()
	if err != nil {
		return TestSnapshot{}, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	_, err = tx.Insert(TestSnapshotsTable).Rows(goqu.Record{
		"id":                     snapshotID,
		"quiz_id":                params.QuizID,
		"created_by":             params.CreatedBy,
		"status":                 TestSnapshotStatusCreated,
		"source_collection_ids":  sourceJSON,
		"question_count":         len(params.Items),
		"created_at":             now,
	}).Executor().Exec()
	if err != nil {
		return TestSnapshot{}, err
	}

	for _, item := range params.Items {
		itemID, err := model.newUUID()
		if err != nil {
			return TestSnapshot{}, err
		}
		record := goqu.Record{
			"id":                    itemID,
			"snapshot_id":           snapshotID,
			"position":              item.Position,
			"collection_id":         item.CollectionID,
			"question_id":           item.Freeze.QuestionID,
			"lineage_id":            item.Freeze.LineageID,
			"revision_number":       item.Freeze.RevisionNumber,
			"question":              item.Freeze.Question,
			"type":                  item.Freeze.Type,
			"options":               item.Freeze.OptionsJSON,
			"answers":               item.Freeze.AnswersJSON,
			"official_answer":       item.Freeze.OfficialAnswerJSON,
			"authoritative_answer":  item.Freeze.AuthoritativeJSON,
			"answer_review_status":  item.Freeze.AnswerReviewStatus,
			"points":                item.Freeze.Points,
			"duration_in_seconds":   item.Freeze.DurationInSeconds,
			"question_media":        item.Freeze.QuestionMedia,
			"options_media":         item.Freeze.OptionsMedia,
			"resource":              item.Freeze.Resource,
			"created_at":            now,
		}
		_, err = tx.Insert(TestSnapshotItemsTable).Rows(record).Executor().Exec()
		if err != nil {
			return TestSnapshot{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return TestSnapshot{}, err
	}
	ok = true
	return model.GetSnapshotByID(params.QuizID, snapshotID)
}

func (model *TestSnapshotModel) ListSnapshotsByQuizID(quizID uuid.UUID) ([]TestSnapshot, error) {
	var rows []TestSnapshot
	err := model.db.From(TestSnapshotsTable).
		Select("id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at").
		Where(goqu.Ex{"quiz_id": quizID}).
		Order(goqu.I("created_at").Desc()).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []TestSnapshot{}
	}
	for i := range rows {
		model.attachSourceIDs(&rows[i])
	}
	return rows, nil
}

func (model *TestSnapshotModel) GetSnapshotByID(quizID, snapshotID uuid.UUID) (TestSnapshot, error) {
	var snapshot TestSnapshot
	found, err := model.db.From(TestSnapshotsTable).
		Select("id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at").
		Where(goqu.Ex{"id": snapshotID, "quiz_id": quizID}).
		ScanStruct(&snapshot)
	if err != nil {
		return TestSnapshot{}, err
	}
	if !found {
		return TestSnapshot{}, ErrTestSnapshotNotFound
	}
	model.attachSourceIDs(&snapshot)

	items, err := model.listItems(snapshotID)
	if err != nil {
		return TestSnapshot{}, err
	}
	snapshot.Items = items
	return snapshot, nil
}

func (model *TestSnapshotModel) ToLearnerView(snapshot TestSnapshot) TestSnapshotLearnerView {
	items := make([]TestSnapshotLearnerItem, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		items = append(items, TestSnapshotLearnerItem{
			Position:          item.Position,
			QuestionID:        item.QuestionID,
			LineageID:         item.LineageID,
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
	return TestSnapshotLearnerView{
		ID:                  snapshot.ID,
		QuizID:              snapshot.QuizID,
		Status:              snapshot.Status,
		SourceCollectionIDs: snapshot.SourceCollectionIDs,
		QuestionCount:       snapshot.QuestionCount,
		CreatedAt:           snapshot.CreatedAt,
		Items:               items,
	}
}

func (model *TestSnapshotModel) listItems(snapshotID uuid.UUID) ([]TestSnapshotItem, error) {
	var items []TestSnapshotItem
	err := model.db.From(TestSnapshotItemsTable).
		Select(
			"id",
			"snapshot_id",
			"position",
			"collection_id",
			"question_id",
			"lineage_id",
			"revision_number",
			"question",
			"type",
			"options",
			"answers",
			"official_answer",
			"authoritative_answer",
			"answer_review_status",
			"points",
			"duration_in_seconds",
			"question_media",
			"options_media",
			"resource",
			"created_at",
		).
		Where(goqu.Ex{"snapshot_id": snapshotID}).
		Order(goqu.I("position").Asc()).
		ScanStructs(&items)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []TestSnapshotItem{}
	}
	for i := range items {
		_ = json.Unmarshal(items[i].OptionsJSON, &items[i].Options)
		if items[i].Options == nil {
			items[i].Options = map[string]string{}
		}
		items[i].Answers, _ = parseAnswerKeys(items[i].AnswersJSON)
		items[i].OfficialAnswer, _ = parseAnswerKeys(items[i].OfficialAnswerJSON)
		items[i].AuthoritativeAnswer, _ = parseAnswerKeys(items[i].AuthoritativeJSON)
	}
	return items, nil
}

func (model *TestSnapshotModel) attachSourceIDs(snapshot *TestSnapshot) {
	ids := []uuid.UUID{}
	if len(snapshot.SourceCollectionJSON) > 0 {
		_ = json.Unmarshal(snapshot.SourceCollectionJSON, &ids)
	}
	if ids == nil {
		ids = []uuid.UUID{}
	}
	snapshot.SourceCollectionIDs = ids
}

// LoadQuestionFreezesByIDs loads live bank rows used only at snapshot creation time.
func (model *TestSnapshotModel) LoadQuestionFreezesByIDs(quizID uuid.UUID, questionIDs []uuid.UUID) (map[uuid.UUID]SnapshotQuestionFreeze, error) {
	result := make(map[uuid.UUID]SnapshotQuestionFreeze, len(questionIDs))
	if len(questionIDs) == 0 {
		return result, nil
	}

	type row struct {
		QuestionID           uuid.UUID `db:"question_id"`
		LineageID            uuid.UUID `db:"lineage_id"`
		RevisionNumber       int       `db:"revision_number"`
		Question             string    `db:"question"`
		Type                 int       `db:"type"`
		Options              []byte    `db:"options"`
		Answers              []byte    `db:"answers"`
		OfficialAnswer       []byte    `db:"official_answer"`
		AuthoritativeAnswer  []byte    `db:"authoritative_answer"`
		AnswerReviewStatus   string    `db:"answer_review_status"`
		Points               *int16    `db:"points"`
		DurationInSeconds    *int      `db:"duration_in_seconds"`
		QuestionMedia        string    `db:"question_media"`
		OptionsMedia         string    `db:"options_media"`
		Resource             *string   `db:"resource"`
	}

	var rows []row
	err := model.db.From(constants.QuizQuestionsTable).
		Select(
			goqu.I(constants.QuestionsTable+".id").As("question_id"),
			goqu.I(constants.QuestionsTable+".lineage_id"),
			goqu.I(constants.QuestionsTable+".revision_number"),
			goqu.I(constants.QuestionsTable+".question"),
			goqu.I(constants.QuestionsTable+".type"),
			goqu.I(constants.QuestionsTable+".options"),
			goqu.I(constants.QuestionsTable+".answers"),
			goqu.I(constants.QuestionsTable+".official_answer"),
			goqu.I(constants.QuestionsTable+".authoritative_answer"),
			goqu.I(constants.QuestionsTable+".answer_review_status"),
			goqu.I(constants.QuestionsTable+".points"),
			goqu.I(constants.QuestionsTable+".duration_in_seconds"),
			goqu.I(constants.QuestionsTable+".question_media"),
			goqu.I(constants.QuestionsTable+".options_media"),
			goqu.I(constants.QuestionsTable+".resource"),
		).
		InnerJoin(
			goqu.T(constants.QuestionsTable),
			goqu.On(goqu.I(constants.QuizQuestionsTable+".question_id").Eq(goqu.I(constants.QuestionsTable+".id"))),
		).
		Where(goqu.Ex{
			constants.QuizQuestionsTable + ".quiz_id":     quizID,
			constants.QuizQuestionsTable + ".question_id": questionIDs,
		}).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}

	for _, item := range rows {
		result[item.QuestionID] = SnapshotQuestionFreeze{
			QuestionID:         item.QuestionID,
			LineageID:          item.LineageID,
			RevisionNumber:     item.RevisionNumber,
			Question:           item.Question,
			Type:               item.Type,
			OptionsJSON:        item.Options,
			AnswersJSON:        item.Answers,
			OfficialAnswerJSON: item.OfficialAnswer,
			AuthoritativeJSON:  item.AuthoritativeAnswer,
			AnswerReviewStatus: item.AnswerReviewStatus,
			Points:             item.Points,
			DurationInSeconds:  item.DurationInSeconds,
			QuestionMedia:      item.QuestionMedia,
			OptionsMedia:       item.OptionsMedia,
			Resource:           item.Resource,
		}
	}
	return result, nil
}
