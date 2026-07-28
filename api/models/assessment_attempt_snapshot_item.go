package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

const AssessmentAttemptSnapshotItemsTable = "assessment_attempt_snapshot_items"

var (
	ErrAttemptSnapshotItemNotFound = errors.New("attempt snapshot item not found")
	ErrAttemptSnapshotItemsEmpty   = errors.New("attempt requires at least one resolved snapshot item")
)

// AssessmentAttemptSnapshotItem is an insert-once copy of a frozen test_snapshot_item
// bound to a specific attempt (EXAM-P5-T04).
type AssessmentAttemptSnapshotItem struct {
	ID                  uuid.UUID         `json:"id" db:"id"`
	AttemptID           uuid.UUID         `json:"attempt_id" db:"attempt_id"`
	SnapshotItemID      uuid.UUID         `json:"snapshot_item_id" db:"snapshot_item_id"`
	Position            int               `json:"position" db:"position"`
	QuestionID          uuid.UUID         `json:"question_id" db:"question_id"`
	LineageID           uuid.UUID         `json:"lineage_id" db:"lineage_id"`
	RevisionNumber      int               `json:"revision_number" db:"revision_number"`
	Question            string            `json:"question" db:"question"`
	Type                int               `json:"type" db:"type"`
	OptionsJSON         []byte            `json:"-" db:"options"`
	AnswersJSON         []byte            `json:"-" db:"answers"`
	OfficialAnswerJSON  []byte            `json:"-" db:"official_answer"`
	AuthoritativeJSON   []byte            `json:"-" db:"authoritative_answer"`
	Options             map[string]string `json:"options" db:"-"`
	Answers             []int             `json:"answers,omitempty" db:"-"`
	OfficialAnswer      []int             `json:"official_answer,omitempty" db:"-"`
	AuthoritativeAnswer []int             `json:"authoritative_answer,omitempty" db:"-"`
	AnswerReviewStatus  string            `json:"answer_review_status" db:"answer_review_status"`
	Points              *int16            `json:"points,omitempty" db:"points"`
	DurationInSeconds   *int              `json:"duration_in_seconds,omitempty" db:"duration_in_seconds"`
	QuestionMedia       string            `json:"question_media,omitempty" db:"question_media"`
	OptionsMedia        string            `json:"options_media,omitempty" db:"options_media"`
	Resource            *string           `json:"resource,omitempty" db:"resource"`
	CreatedAt           time.Time         `json:"created_at" db:"created_at"`
}

type CreateAttemptSnapshotItemParams struct {
	SnapshotItemID      uuid.UUID
	Position            int
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

type AssessmentAttemptSnapshotItemModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitAssessmentAttemptSnapshotItemModel(goquDB *goqu.Database) *AssessmentAttemptSnapshotItemModel {
	return &AssessmentAttemptSnapshotItemModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

func (model *AssessmentAttemptSnapshotItemModel) SetUUIDGenerator(fn func() (uuid.UUID, error)) {
	if fn != nil {
		model.newUUID = fn
	}
}

func (model *AssessmentAttemptSnapshotItemModel) InsertItemsTx(tx *goqu.TxDatabase, attemptID uuid.UUID, items []CreateAttemptSnapshotItemParams) error {
	if tx == nil {
		return errors.New("attempt snapshot item insert requires a transaction")
	}
	if len(items) == 0 {
		return ErrAttemptSnapshotItemsEmpty
	}
	now := time.Now().UTC()
	for _, item := range items {
		itemID, err := model.newUUID()
		if err != nil {
			return err
		}
		_, err = tx.Insert(AssessmentAttemptSnapshotItemsTable).Rows(goqu.Record{
			"id":                    itemID,
			"attempt_id":            attemptID,
			"snapshot_item_id":      item.SnapshotItemID,
			"position":              item.Position,
			"question_id":           item.QuestionID,
			"lineage_id":            item.LineageID,
			"revision_number":       item.RevisionNumber,
			"question":              item.Question,
			"type":                  item.Type,
			"options":               item.OptionsJSON,
			"answers":               item.AnswersJSON,
			"official_answer":       item.OfficialAnswerJSON,
			"authoritative_answer":  item.AuthoritativeJSON,
			"answer_review_status":  item.AnswerReviewStatus,
			"points":                item.Points,
			"duration_in_seconds":   item.DurationInSeconds,
			"question_media":        item.QuestionMedia,
			"options_media":         item.OptionsMedia,
			"resource":              item.Resource,
			"created_at":            now,
		}).Executor().Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func (model *AssessmentAttemptSnapshotItemModel) ListByAttemptID(attemptID uuid.UUID) ([]AssessmentAttemptSnapshotItem, error) {
	var rows []AssessmentAttemptSnapshotItem
	err := model.db.From(AssessmentAttemptSnapshotItemsTable).
		Select(
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		).
		Where(goqu.Ex{"attempt_id": attemptID}).
		Order(goqu.I("position").Asc()).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []AssessmentAttemptSnapshotItem{}
	}
	for i := range rows {
		if err := decodeAttemptSnapshotItem(&rows[i]); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (model *AssessmentAttemptSnapshotItemModel) GetByAttemptAndQuestion(attemptID, questionID uuid.UUID) (AssessmentAttemptSnapshotItem, error) {
	var row AssessmentAttemptSnapshotItem
	found, err := model.db.From(AssessmentAttemptSnapshotItemsTable).
		Select(
			"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		).
		Where(goqu.Ex{"attempt_id": attemptID, "question_id": questionID}).
		ScanStruct(&row)
	if err != nil {
		return AssessmentAttemptSnapshotItem{}, err
	}
	if !found {
		return AssessmentAttemptSnapshotItem{}, ErrAttemptSnapshotItemNotFound
	}
	if err := decodeAttemptSnapshotItem(&row); err != nil {
		return AssessmentAttemptSnapshotItem{}, err
	}
	return row, nil
}

func decodeAttemptSnapshotItem(item *AssessmentAttemptSnapshotItem) error {
	if len(item.OptionsJSON) > 0 {
		if err := json.Unmarshal(item.OptionsJSON, &item.Options); err != nil {
			return err
		}
	} else {
		item.Options = map[string]string{}
	}
	if len(item.AnswersJSON) > 0 {
		if err := json.Unmarshal(item.AnswersJSON, &item.Answers); err != nil {
			return err
		}
	}
	if len(item.OfficialAnswerJSON) > 0 {
		if err := json.Unmarshal(item.OfficialAnswerJSON, &item.OfficialAnswer); err != nil {
			return err
		}
	}
	if len(item.AuthoritativeJSON) > 0 {
		if err := json.Unmarshal(item.AuthoritativeJSON, &item.AuthoritativeAnswer); err != nil {
			return err
		}
	}
	return nil
}

// ToLearnerItem projects an attempt-linked freeze without answer keys.
func (item AssessmentAttemptSnapshotItem) ToLearnerItem() TestSnapshotLearnerItem {
	return TestSnapshotLearnerItem{
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
	}
}

// SnapshotItemFromTestFreeze builds create params from a shared test snapshot item.
func SnapshotItemFromTestFreeze(item TestSnapshotItem) CreateAttemptSnapshotItemParams {
	return CreateAttemptSnapshotItemParams{
		SnapshotItemID:     item.ID,
		Position:           item.Position,
		QuestionID:         item.QuestionID,
		LineageID:          item.LineageID,
		RevisionNumber:     item.RevisionNumber,
		Question:           item.Question,
		Type:               item.Type,
		OptionsJSON:        item.OptionsJSON,
		AnswersJSON:        item.AnswersJSON,
		OfficialAnswerJSON: item.OfficialAnswerJSON,
		AuthoritativeJSON:  item.AuthoritativeJSON,
		AnswerReviewStatus: item.AnswerReviewStatus,
		Points:             item.Points,
		DurationInSeconds:  item.DurationInSeconds,
		QuestionMedia:      item.QuestionMedia,
		OptionsMedia:       item.OptionsMedia,
		Resource:           item.Resource,
	}
}
