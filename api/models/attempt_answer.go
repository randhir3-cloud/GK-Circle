package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

const AttemptAnswersTable = "attempt_answers"
const attemptAnswersUniqueConstraint = "attempt_answers_unique_per_question"

var (
	ErrAttemptAnswerNotFound              = errors.New("attempt answer not found")
	ErrAttemptAnswerInvalidOptions        = errors.New("attempt answer blocked: selected options are invalid")
	ErrAttemptAnswerInvalidOptionRef      = errors.New("attempt answer blocked: option is not in the frozen snapshot")
	ErrAttemptAnswerCardinality           = errors.New("attempt answer blocked: selected option cardinality is invalid for question type")
	ErrAttemptAnswerQuestionNotInSnapshot = errors.New("attempt answer blocked: question is not in the attempt snapshot")
	ErrAttemptAnswerNotInProgress         = errors.New("attempt answer blocked: attempt is not in progress")
	ErrAttemptAnswerClientScoreForbidden  = errors.New("attempt answer blocked: client must not supply score or correctness")
)

// AttemptAnswer is the per-question answer row for a self-paced attempt.
type AttemptAnswer struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	AttemptID        uuid.UUID       `json:"attempt_id" db:"attempt_id"`
	QuestionID       uuid.UUID       `json:"question_id" db:"question_id"`
	SelectedOptions  []byte          `json:"selected_options,omitempty" db:"selected_options"`
	IsMarkedReview   bool            `json:"is_marked_review" db:"is_marked_review"`
	AnsweredAt       sql.NullTime    `json:"answered_at,omitempty" db:"answered_at"`
	TimeTakenSeconds sql.NullInt64   `json:"time_taken_seconds,omitempty" db:"time_taken_seconds"`
	Score            sql.NullFloat64 `json:"score,omitempty" db:"score"`
	IsCorrect        sql.NullBool    `json:"is_correct,omitempty" db:"is_correct"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// UpsertAttemptAnswerParams writes learner selection without score/correctness.
type UpsertAttemptAnswerParams struct {
	AttemptID        uuid.UUID
	QuestionID       uuid.UUID
	SelectedOptions  []int // empty clears the answer
	ClearAnswer      bool
	IsMarkedReview   bool
	TimeTakenSeconds *int
}

type AttemptAnswerModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitAttemptAnswerModel(goquDB *goqu.Database) *AttemptAnswerModel {
	return &AttemptAnswerModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

func (model *AttemptAnswerModel) SetUUIDGenerator(fn func() (uuid.UUID, error)) {
	if fn != nil {
		model.newUUID = fn
	}
}

func (model *AttemptAnswerModel) ListByAttemptID(attemptID uuid.UUID) ([]AttemptAnswer, error) {
	var rows []AttemptAnswer
	err := model.db.From(AttemptAnswersTable).
		Select(
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		).
		Where(goqu.Ex{"attempt_id": attemptID}).
		Order(goqu.I("created_at").Asc()).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []AttemptAnswer{}
	}
	return rows, nil
}

func (model *AttemptAnswerModel) GetByAttemptAndQuestion(attemptID, questionID uuid.UUID) (AttemptAnswer, error) {
	var row AttemptAnswer
	found, err := model.db.From(AttemptAnswersTable).
		Select(
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		).
		Where(goqu.Ex{"attempt_id": attemptID, "question_id": questionID}).
		ScanStruct(&row)
	if err != nil {
		return AttemptAnswer{}, err
	}
	if !found {
		return AttemptAnswer{}, ErrAttemptAnswerNotFound
	}
	return row, nil
}

// UpsertAnswer inserts or updates a single answer row. score/is_correct are never set.
func (model *AttemptAnswerModel) UpsertAnswer(tx *goqu.TxDatabase, params UpsertAttemptAnswerParams) (AttemptAnswer, error) {
	if tx == nil {
		return AttemptAnswer{}, errors.New("attempt answer upsert requires a transaction")
	}

	var selectedJSON []byte
	var answeredAt interface{}
	now := time.Now().UTC()
	if params.ClearAnswer || len(params.SelectedOptions) == 0 {
		selectedJSON = nil
		answeredAt = nil
	} else {
		normalized := uniqueSortedInts(params.SelectedOptions)
		payload, err := json.Marshal(normalized)
		if err != nil {
			return AttemptAnswer{}, err
		}
		selectedJSON = payload
		answeredAt = now
	}

	var timeTaken interface{}
	if params.TimeTakenSeconds != nil {
		timeTaken = *params.TimeTakenSeconds
	}

	existing, err := model.getByAttemptAndQuestionTx(tx, params.AttemptID, params.QuestionID)
	if err != nil && !errors.Is(err, ErrAttemptAnswerNotFound) {
		return AttemptAnswer{}, err
	}

	if errors.Is(err, ErrAttemptAnswerNotFound) {
		answerID, idErr := model.newUUID()
		if idErr != nil {
			return AttemptAnswer{}, idErr
		}
		_, err = tx.Insert(AttemptAnswersTable).Rows(goqu.Record{
			"id":                 answerID,
			"attempt_id":         params.AttemptID,
			"question_id":        params.QuestionID,
			"selected_options":   selectedJSON,
			"is_marked_review":   params.IsMarkedReview,
			"answered_at":        answeredAt,
			"time_taken_seconds": timeTaken,
			"score":              nil,
			"is_correct":         nil,
			"created_at":         now,
			"updated_at":         now,
		}).Executor().Exec()
		if err != nil {
			if isAttemptAnswerUniqueConflict(err) {
				existing, getErr := model.getByAttemptAndQuestionTx(tx, params.AttemptID, params.QuestionID)
				if getErr != nil {
					return AttemptAnswer{}, getErr
				}
				return model.updateAnswerTx(tx, existing.ID, params.AttemptID, params.QuestionID, selectedJSON, answeredAt, timeTaken, params.IsMarkedReview, now)
			}
			return AttemptAnswer{}, err
		}
		return model.getByAttemptAndQuestionTx(tx, params.AttemptID, params.QuestionID)
	}

	return model.updateAnswerTx(tx, existing.ID, params.AttemptID, params.QuestionID, selectedJSON, answeredAt, timeTaken, params.IsMarkedReview, now)
}

func (model *AttemptAnswerModel) updateAnswerTx(
	tx *goqu.TxDatabase,
	answerID, attemptID, questionID uuid.UUID,
	selectedJSON []byte,
	answeredAt, timeTaken interface{},
	isMarkedReview bool,
	now time.Time,
) (AttemptAnswer, error) {
	record := goqu.Record{
		"selected_options": selectedJSON,
		"is_marked_review": isMarkedReview,
		"answered_at":      answeredAt,
		"updated_at":       now,
		"score":            nil,
		"is_correct":       nil,
	}
	if timeTaken != nil {
		record["time_taken_seconds"] = timeTaken
	}
	_, err := tx.Update(AttemptAnswersTable).
		Set(record).
		Where(goqu.Ex{"id": answerID, "attempt_id": attemptID, "question_id": questionID}).
		Executor().Exec()
	if err != nil {
		return AttemptAnswer{}, err
	}
	return model.getByAttemptAndQuestionTx(tx, attemptID, questionID)
}

func (model *AttemptAnswerModel) getByAttemptAndQuestionTx(tx *goqu.TxDatabase, attemptID, questionID uuid.UUID) (AttemptAnswer, error) {
	var row AttemptAnswer
	found, err := tx.From(AttemptAnswersTable).
		Select(
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		).
		Where(goqu.Ex{"attempt_id": attemptID, "question_id": questionID}).
		ScanStruct(&row)
	if err != nil {
		return AttemptAnswer{}, err
	}
	if !found {
		return AttemptAnswer{}, ErrAttemptAnswerNotFound
	}
	return row, nil
}

// ApplyScoreOutcomeParams writes scoring fields for one question after submit.
type ApplyScoreOutcomeParams struct {
	AttemptID       uuid.UUID
	QuestionID      uuid.UUID
	SelectedOptions []int
	IsMarkedReview  bool
	IsCorrect       *bool
	Score           float64
	AnsweredAt      *time.Time
	TimeTakenSeconds *int
}

// ApplyScoreOutcomeTx upserts score/is_correct for a question in the submit transaction.
func (model *AttemptAnswerModel) ApplyScoreOutcomeTx(tx *goqu.TxDatabase, params ApplyScoreOutcomeParams) error {
	if tx == nil {
		return errors.New("attempt answer score apply requires a transaction")
	}
	now := time.Now().UTC()
	var selectedJSON []byte
	if len(params.SelectedOptions) > 0 {
		payload, err := json.Marshal(uniqueSortedInts(params.SelectedOptions))
		if err != nil {
			return err
		}
		selectedJSON = payload
	}
	var answeredAt interface{}
	if params.AnsweredAt != nil {
		answeredAt = *params.AnsweredAt
	}
	var timeTaken interface{}
	if params.TimeTakenSeconds != nil {
		timeTaken = *params.TimeTakenSeconds
	}
	var isCorrect interface{}
	if params.IsCorrect != nil {
		isCorrect = *params.IsCorrect
	}

	existing, err := model.getByAttemptAndQuestionTx(tx, params.AttemptID, params.QuestionID)
	if err != nil && !errors.Is(err, ErrAttemptAnswerNotFound) {
		return err
	}
	if errors.Is(err, ErrAttemptAnswerNotFound) {
		answerID, idErr := model.newUUID()
		if idErr != nil {
			return idErr
		}
		_, err = tx.Insert(AttemptAnswersTable).Rows(goqu.Record{
			"id":                 answerID,
			"attempt_id":         params.AttemptID,
			"question_id":        params.QuestionID,
			"selected_options":   selectedJSON,
			"is_marked_review":   params.IsMarkedReview,
			"answered_at":        answeredAt,
			"time_taken_seconds": timeTaken,
			"score":              params.Score,
			"is_correct":         isCorrect,
			"created_at":         now,
			"updated_at":         now,
		}).Executor().Exec()
		return err
	}

	_, err = tx.Update(AttemptAnswersTable).
		Set(goqu.Record{
			"selected_options":   selectedJSON,
			"is_marked_review":   params.IsMarkedReview,
			"answered_at":        answeredAt,
			"time_taken_seconds": timeTaken,
			"score":              params.Score,
			"is_correct":         isCorrect,
			"updated_at":         now,
		}).
		Where(goqu.Ex{"id": existing.ID, "attempt_id": params.AttemptID, "question_id": params.QuestionID}).
		Executor().Exec()
	return err
}

func isAttemptAnswerUniqueConflict(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		return false
	}
	return pqErr.Constraint == attemptAnswersUniqueConstraint
}

// ValidateSelectedOptionsAgainstSnapshot checks option refs and cardinality for frozen items.
func ValidateSelectedOptionsAgainstSnapshot(itemType int, options map[string]string, selected []int, clear bool) error {
	if clear || len(selected) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(selected))
	for _, opt := range selected {
		if opt <= 0 {
			return ErrAttemptAnswerInvalidOptions
		}
		if _, dup := seen[opt]; dup {
			return ErrAttemptAnswerInvalidOptions
		}
		seen[opt] = struct{}{}
		key := strconv.Itoa(opt)
		if _, ok := options[key]; !ok {
			return ErrAttemptAnswerInvalidOptionRef
		}
	}
	switch itemType {
	case constants.SingleAnswer:
		if len(selected) != 1 {
			return ErrAttemptAnswerCardinality
		}
	case constants.Survey:
		if len(selected) < 1 || len(selected) > len(options) {
			return ErrAttemptAnswerCardinality
		}
	default:
		return fmt.Errorf("%w: unsupported question type %d", ErrAttemptAnswerCardinality, itemType)
	}
	return nil
}

func DecodeSelectedOptions(raw []byte) ([]int, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var selected []int
	if err := json.Unmarshal(raw, &selected); err != nil {
		return nil, ErrAttemptAnswerInvalidOptions
	}
	return selected, nil
}

func uniqueSortedInts(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	out := make([]int, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}
