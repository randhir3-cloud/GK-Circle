package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

const QuestionCollectionsTable = "question_collections"
const QuestionCollectionMembersTable = "question_collection_members"

type QuestionCollectionKind string

const (
	QuestionCollectionKindStatic  QuestionCollectionKind = "STATIC"
	QuestionCollectionKindDynamic QuestionCollectionKind = "DYNAMIC"
)

const (
	CollectionResolutionStatusResolved         = "RESOLVED"
	CollectionResolutionStatusMetadataPending  = "METADATA_PENDING"
	CollectionResolutionStatusEmptyFilterAll   = "ALL_QUIZ_QUESTIONS"
)

var (
	ErrQuestionCollectionNotFound       = errors.New("question collection not found")
	ErrQuestionCollectionTitleRequired  = errors.New("question collection title is required")
	ErrQuestionCollectionKindInvalid    = errors.New("question collection kind is invalid")
	ErrQuestionCollectionKindMismatch   = errors.New("question collection kind does not support this operation")
	ErrQuestionCollectionFilterRequired = errors.New("dynamic question collection requires filter_json")
	ErrQuestionCollectionFilterInvalid  = errors.New("dynamic question collection filter is invalid")
	ErrQuestionCollectionMemberInvalid  = errors.New("question collection member is not linked to quiz")
	ErrQuestionCollectionMemberDuplicate = errors.New("question collection member is duplicated")
	ErrQuestionCollectionUpdateRequired = errors.New("question collection update requires at least one field")
)

// CollectionDynamicFilter stores DYNAMIC collection criteria per ADR-024 §3.
// Resolution against question taxonomy ships when bank metadata is available (EXAM-P10).
type CollectionDynamicFilter struct {
	Subject    *string `json:"subject,omitempty"`
	Topic      *string `json:"topic,omitempty"`
	Year       *int    `json:"year,omitempty"`
	Difficulty *string `json:"difficulty,omitempty"`
	PYQStatus  *bool   `json:"pyq_status,omitempty"`
}

func (filter CollectionDynamicFilter) HasMetadataCriteria() bool {
	return filter.Subject != nil ||
		filter.Topic != nil ||
		filter.Year != nil ||
		filter.Difficulty != nil ||
		filter.PYQStatus != nil
}

func ParseCollectionDynamicFilter(raw json.RawMessage) (CollectionDynamicFilter, error) {
	if len(raw) == 0 {
		return CollectionDynamicFilter{}, ErrQuestionCollectionFilterInvalid
	}
	var filter CollectionDynamicFilter
	if err := json.Unmarshal(raw, &filter); err != nil {
		return CollectionDynamicFilter{}, ErrQuestionCollectionFilterInvalid
	}
	return filter, nil
}

func ValidateCollectionDynamicFilter(raw json.RawMessage) error {
	_, err := ParseCollectionDynamicFilter(raw)
	return err
}

type QuestionCollection struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	QuizID     uuid.UUID              `json:"quiz_id" db:"quiz_id"`
	Title      string                 `json:"title" db:"title"`
	Kind       QuestionCollectionKind `json:"kind" db:"kind"`
	Position   int                    `json:"position" db:"position"`
	FilterJSON []byte                 `json:"-" db:"filter_json"`
	Filter     *CollectionDynamicFilter `json:"filter,omitempty"`
	CreatedBy  string                 `json:"created_by" db:"created_by"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
	Members    []QuestionCollectionMember `json:"members,omitempty"`
}

type QuestionCollectionMember struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CollectionID uuid.UUID `json:"collection_id" db:"collection_id"`
	QuestionID   uuid.UUID `json:"question_id" db:"question_id"`
	Position     int       `json:"position" db:"position"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CollectionResolution struct {
	CollectionID     uuid.UUID `json:"collection_id"`
	Kind             QuestionCollectionKind `json:"kind"`
	QuestionIDs      []uuid.UUID `json:"question_ids"`
	ResolutionStatus string    `json:"resolution_status"`
	Message          string    `json:"message,omitempty"`
}

type CreateQuestionCollectionParams struct {
	QuizID    uuid.UUID
	Title     string
	Kind      QuestionCollectionKind
	Position  int
	FilterJSON json.RawMessage
	CreatedBy string
}

type UpdateQuestionCollectionParams struct {
	Title      *string
	Position   *int
	FilterJSON json.RawMessage
	HasFilter  bool
}

type QuestionCollectionModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitQuestionCollectionModel(goquDB *goqu.Database) *QuestionCollectionModel {
	return &QuestionCollectionModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

var questionCollectionColumns = []interface{}{
	"id",
	"quiz_id",
	"title",
	"kind",
	"position",
	"filter_json",
	"created_by",
	"created_at",
	"updated_at",
}

func (model *QuestionCollectionModel) CreateCollection(params CreateQuestionCollectionParams) (QuestionCollection, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return QuestionCollection{}, ErrQuestionCollectionTitleRequired
	}
	if params.Kind != QuestionCollectionKindStatic && params.Kind != QuestionCollectionKindDynamic {
		return QuestionCollection{}, ErrQuestionCollectionKindInvalid
	}

	var filterJSON interface{}
	switch params.Kind {
	case QuestionCollectionKindStatic:
		if len(params.FilterJSON) > 0 {
			return QuestionCollection{}, ErrQuestionCollectionFilterInvalid
		}
		filterJSON = nil
	case QuestionCollectionKindDynamic:
		if len(params.FilterJSON) == 0 {
			return QuestionCollection{}, ErrQuestionCollectionFilterRequired
		}
		if err := ValidateCollectionDynamicFilter(params.FilterJSON); err != nil {
			return QuestionCollection{}, err
		}
		filterJSON = params.FilterJSON
	}

	id, err := model.newUUID()
	if err != nil {
		return QuestionCollection{}, err
	}

	now := time.Now().UTC()
	position := params.Position
	if position < 0 {
		position = 0
	}

	record := goqu.Record{
		"id":          id,
		"quiz_id":     params.QuizID,
		"title":       title,
		"kind":        params.Kind,
		"position":    position,
		"filter_json": filterJSON,
		"created_by":  strings.TrimSpace(params.CreatedBy),
		"created_at":  now,
		"updated_at":  now,
	}

	_, err = model.db.Insert(QuestionCollectionsTable).Rows(record).Executor().Exec()
	if err != nil {
		return QuestionCollection{}, err
	}

	return model.GetCollectionByID(params.QuizID, id)
}

func (model *QuestionCollectionModel) ListCollectionsByQuizID(quizID uuid.UUID) ([]QuestionCollection, error) {
	var rows []QuestionCollection
	err := model.db.From(QuestionCollectionsTable).
		Select(questionCollectionColumns...).
		Where(goqu.Ex{"quiz_id": quizID}).
		Order(goqu.I("position").Asc(), goqu.I("created_at").Asc()).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []QuestionCollection{}
	}
	for i := range rows {
		model.attachFilter(&rows[i])
	}
	return rows, nil
}

func (model *QuestionCollectionModel) GetCollectionByID(quizID, collectionID uuid.UUID) (QuestionCollection, error) {
	var collection QuestionCollection
	found, err := model.db.From(QuestionCollectionsTable).
		Select(questionCollectionColumns...).
		Where(goqu.Ex{"id": collectionID, "quiz_id": quizID}).
		ScanStruct(&collection)
	if err != nil {
		return QuestionCollection{}, err
	}
	if !found {
		return QuestionCollection{}, ErrQuestionCollectionNotFound
	}
	model.attachFilter(&collection)
	return collection, nil
}

func (model *QuestionCollectionModel) GetCollectionWithMembers(quizID, collectionID uuid.UUID) (QuestionCollection, error) {
	collection, err := model.GetCollectionByID(quizID, collectionID)
	if err != nil {
		return QuestionCollection{}, err
	}
	if collection.Kind != QuestionCollectionKindStatic {
		return collection, nil
	}
	members, err := model.listMembers(collectionID)
	if err != nil {
		return QuestionCollection{}, err
	}
	collection.Members = members
	return collection, nil
}

func (model *QuestionCollectionModel) UpdateCollection(quizID, collectionID uuid.UUID, params UpdateQuestionCollectionParams) (QuestionCollection, error) {
	collection, err := model.GetCollectionByID(quizID, collectionID)
	if err != nil {
		return QuestionCollection{}, err
	}

	record := goqu.Record{"updated_at": time.Now().UTC()}
	hasChange := false

	if params.Title != nil {
		title := strings.TrimSpace(*params.Title)
		if title == "" {
			return QuestionCollection{}, ErrQuestionCollectionTitleRequired
		}
		record["title"] = title
		hasChange = true
	}
	if params.Position != nil {
		record["position"] = *params.Position
		hasChange = true
	}
	if params.HasFilter {
		if collection.Kind != QuestionCollectionKindDynamic {
			return QuestionCollection{}, ErrQuestionCollectionKindMismatch
		}
		if len(params.FilterJSON) == 0 {
			return QuestionCollection{}, ErrQuestionCollectionFilterRequired
		}
		if err := ValidateCollectionDynamicFilter(params.FilterJSON); err != nil {
			return QuestionCollection{}, err
		}
		record["filter_json"] = params.FilterJSON
		hasChange = true
	}
	if !hasChange {
		return QuestionCollection{}, ErrQuestionCollectionUpdateRequired
	}

	_, err = model.db.Update(QuestionCollectionsTable).
		Set(record).
		Where(goqu.Ex{"id": collectionID, "quiz_id": quizID}).
		Executor().
		Exec()
	if err != nil {
		return QuestionCollection{}, err
	}
	return model.GetCollectionWithMembers(quizID, collectionID)
}

func (model *QuestionCollectionModel) DeleteCollection(quizID, collectionID uuid.UUID) error {
	result, err := model.db.Delete(QuestionCollectionsTable).
		Where(goqu.Ex{"id": collectionID, "quiz_id": quizID}).
		Executor().
		Exec()
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrQuestionCollectionNotFound
	}
	return nil
}

func (model *QuestionCollectionModel) ReplaceStaticMembers(quizID, collectionID uuid.UUID, questionIDs []uuid.UUID) ([]QuestionCollectionMember, error) {
	collection, err := model.GetCollectionByID(quizID, collectionID)
	if err != nil {
		return nil, err
	}
	if collection.Kind != QuestionCollectionKindStatic {
		return nil, ErrQuestionCollectionKindMismatch
	}

	seen := make(map[uuid.UUID]struct{}, len(questionIDs))
	for _, questionID := range questionIDs {
		if _, ok := seen[questionID]; ok {
			return nil, ErrQuestionCollectionMemberDuplicate
		}
		seen[questionID] = struct{}{}
	}

	if len(questionIDs) > 0 {
		if err := model.assertQuestionsLinkedToQuiz(quizID, questionIDs); err != nil {
			return nil, err
		}
	}

	tx, err := model.db.Begin()
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Delete(QuestionCollectionMembersTable).
		Where(goqu.Ex{"collection_id": collectionID}).
		Executor().
		Exec()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if len(questionIDs) > 0 {
		for index, questionID := range questionIDs {
			memberID, err := model.newUUID()
			if err != nil {
				return nil, err
			}
			_, err = tx.Insert(QuestionCollectionMembersTable).Rows(goqu.Record{
				"id":            memberID,
				"collection_id": collectionID,
				"question_id":   questionID,
				"position":      index,
				"created_at":    now,
			}).Executor().Exec()
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	ok = true
	return model.listMembers(collectionID)
}

func (model *QuestionCollectionModel) ResolveCollection(quizID, collectionID uuid.UUID) (CollectionResolution, error) {
	collection, err := model.GetCollectionWithMembers(quizID, collectionID)
	if err != nil {
		return CollectionResolution{}, err
	}

	resolution := CollectionResolution{
		CollectionID: collection.ID,
		Kind:         collection.Kind,
		QuestionIDs:  []uuid.UUID{},
	}

	switch collection.Kind {
	case QuestionCollectionKindStatic:
		resolution.ResolutionStatus = CollectionResolutionStatusResolved
		for _, member := range collection.Members {
			resolution.QuestionIDs = append(resolution.QuestionIDs, member.QuestionID)
		}
		return resolution, nil
	case QuestionCollectionKindDynamic:
		filter, err := ParseCollectionDynamicFilter(collection.FilterJSON)
		if err != nil {
			return CollectionResolution{}, err
		}
		if filter.HasMetadataCriteria() {
			resolution.ResolutionStatus = CollectionResolutionStatusMetadataPending
			resolution.Message = "question bank taxonomy metadata is not yet available; filter criteria are stored but not resolved"
			return resolution, nil
		}
		questionIDs, err := model.listQuizQuestionIDs(quizID)
		if err != nil {
			return CollectionResolution{}, err
		}
		resolution.QuestionIDs = questionIDs
		resolution.ResolutionStatus = CollectionResolutionStatusEmptyFilterAll
		resolution.Message = "empty dynamic filter resolves to all quiz-linked questions"
		return resolution, nil
	default:
		return CollectionResolution{}, ErrQuestionCollectionKindInvalid
	}
}

func (model *QuestionCollectionModel) listMembers(collectionID uuid.UUID) ([]QuestionCollectionMember, error) {
	var members []QuestionCollectionMember
	err := model.db.From(QuestionCollectionMembersTable).
		Select("id", "collection_id", "question_id", "position", "created_at").
		Where(goqu.Ex{"collection_id": collectionID}).
		Order(goqu.I("position").Asc()).
		ScanStructs(&members)
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []QuestionCollectionMember{}
	}
	return members, nil
}

func (model *QuestionCollectionModel) listQuizQuestionIDs(quizID uuid.UUID) ([]uuid.UUID, error) {
	type row struct {
		QuestionID uuid.UUID `db:"question_id"`
	}
	var rows []row
	err := model.db.From(constants.QuizQuestionsTable).
		Select("question_id").
		Where(goqu.Ex{"quiz_id": quizID}).
		Order(goqu.I("created_at").Asc()).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, item := range rows {
		ids = append(ids, item.QuestionID)
	}
	return ids, nil
}

func (model *QuestionCollectionModel) assertQuestionsLinkedToQuiz(quizID uuid.UUID, questionIDs []uuid.UUID) error {
	if len(questionIDs) == 0 {
		return nil
	}
	type row struct {
		QuestionID uuid.UUID `db:"question_id"`
	}
	var rows []row
	err := model.db.From(constants.QuizQuestionsTable).
		Select("question_id").
		Where(goqu.Ex{
			"quiz_id":     quizID,
			"question_id": questionIDs,
		}).
		ScanStructs(&rows)
	if err != nil {
		return err
	}
	if len(rows) != len(questionIDs) {
		return ErrQuestionCollectionMemberInvalid
	}
	return nil
}

func (model *QuestionCollectionModel) attachFilter(collection *QuestionCollection) {
	if collection.Kind != QuestionCollectionKindDynamic || len(collection.FilterJSON) == 0 {
		collection.Filter = nil
		return
	}
	filter, err := ParseCollectionDynamicFilter(collection.FilterJSON)
	if err != nil {
		collection.Filter = nil
		return
	}
	collection.Filter = &filter
}
