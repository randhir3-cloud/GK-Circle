package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const AssessmentAttemptsTable = "assessment_attempts"

const (
	AttemptStatusInProgress    = "IN_PROGRESS"
	AttemptStatusSubmitted     = "SUBMITTED"
	AttemptStatusAutoSubmitted = "AUTO_SUBMITTED"
	AttemptStatusAbandoned     = "ABANDONED"
)

const assessmentAttemptsOneActiveConstraint = "assessment_attempts_one_active_uidx"

var (
	ErrAssessmentAttemptNotFound         = errors.New("assessment attempt not found")
	ErrAssessmentAttemptSnapshotRequired = errors.New("assessment attempt requires a test snapshot")
	ErrAssessmentAttemptEmptySnapshot    = errors.New("assessment attempt blocked: snapshot has no questions")
	ErrAssessmentAttemptForeignSnapshot  = errors.New("assessment attempt blocked: snapshot does not belong to quiz")
	ErrAssessmentAttemptMaxReached       = errors.New("assessment attempt blocked: max attempts reached")
	ErrAssessmentAttemptNotSelfPaced     = errors.New("assessment attempt blocked: quiz is not self-paced")
	ErrAssessmentAttemptNotEntitled      = errors.New("assessment attempt blocked: not entitled to this quiz")
	ErrAssessmentAttemptQuizNotPublished = errors.New("assessment attempt blocked: quiz is not published")
	ErrAssessmentAttemptOwnerMismatch    = errors.New("assessment attempt blocked: ownership mismatch")
	ErrAssessmentAttemptAlreadyTerminal  = errors.New("assessment attempt blocked: attempt is already terminal")
	ErrAssessmentAttemptSubmitConflict   = errors.New("assessment attempt blocked: submit conflict")
	ErrAssessmentAttemptNotSubmitted     = errors.New("assessment attempt blocked: attempt is not submitted")
)

// AssessmentAttempt is a self-paced attempt bound to an immutable test snapshot.
type AssessmentAttempt struct {
	ID                       uuid.UUID       `json:"id" db:"id"`
	QuizID                   uuid.UUID       `json:"quiz_id" db:"quiz_id"`
	UserID                   string          `json:"user_id" db:"user_id"`
	TestSnapshotID           uuid.UUID       `json:"test_snapshot_id" db:"test_snapshot_id"`
	AttemptNumber            int             `json:"attempt_number" db:"attempt_number"`
	Status                   string          `json:"status" db:"status"`
	QuestionOrderJSON        []byte          `json:"-" db:"question_order"`
	QuestionOrder            []uuid.UUID     `json:"question_order" db:"-"`
	NegativeMarksPerQuestion float64         `json:"negative_marks_per_question" db:"negative_marks_per_question"`
	ExpectedMaxScore         sql.NullFloat64 `json:"expected_max_score,omitempty" db:"expected_max_score"`
	StartedAt                time.Time       `json:"started_at" db:"started_at"`
	SubmittedAt              sql.NullTime    `json:"submitted_at,omitempty" db:"submitted_at"`
	ExpiresAt                sql.NullTime    `json:"expires_at,omitempty" db:"expires_at"`
	TotalScore               sql.NullFloat64 `json:"total_score,omitempty" db:"total_score"`
	MaxScore                 sql.NullFloat64 `json:"max_score,omitempty" db:"max_score"`
	TimeTakenSeconds         sql.NullInt64   `json:"time_taken_seconds,omitempty" db:"time_taken_seconds"`
	CreatedAt                time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at" db:"updated_at"`
}

type CreateAssessmentAttemptParams struct {
	QuizID                   uuid.UUID
	UserID                   string
	TestSnapshotID           uuid.UUID
	AttemptNumber            int
	QuestionOrder            []uuid.UUID
	NegativeMarksPerQuestion float64
	ExpectedMaxScore         float64
	StartedAt                time.Time
	ExpiresAt                *time.Time
	SnapshotItems            []CreateAttemptSnapshotItemParams
	// BeforeCommit runs inside the create transaction after attempt + snapshot items are inserted.
	BeforeCommit func(tx *goqu.TxDatabase, attemptID uuid.UUID) error
}

type AssessmentAttemptModel struct {
	db      *goqu.Database
	newUUID func() (uuid.UUID, error)
}

func InitAssessmentAttemptModel(goquDB *goqu.Database) *AssessmentAttemptModel {
	return &AssessmentAttemptModel{
		db:      goquDB,
		newUUID: uuid.NewUUID,
	}
}

func (model *AssessmentAttemptModel) SetUUIDGenerator(fn func() (uuid.UUID, error)) {
	if fn != nil {
		model.newUUID = fn
	}
}

func (model *AssessmentAttemptModel) GetByID(quizID, attemptID uuid.UUID) (AssessmentAttempt, error) {
	var row AssessmentAttempt
	found, err := model.db.From(AssessmentAttemptsTable).
		Select(
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		).
		Where(goqu.Ex{"id": attemptID, "quiz_id": quizID}).
		ScanStruct(&row)
	if err != nil {
		return AssessmentAttempt{}, err
	}
	if !found {
		return AssessmentAttempt{}, ErrAssessmentAttemptNotFound
	}
	if err := decodeAttemptQuestionOrder(&row); err != nil {
		return AssessmentAttempt{}, err
	}
	return row, nil
}

func (model *AssessmentAttemptModel) GetInProgress(quizID uuid.UUID, userID string) (AssessmentAttempt, bool, error) {
	var row AssessmentAttempt
	found, err := model.db.From(AssessmentAttemptsTable).
		Select(
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		).
		Where(goqu.Ex{
			"quiz_id": quizID,
			"user_id": userID,
			"status":  AttemptStatusInProgress,
		}).
		ScanStruct(&row)
	if err != nil {
		return AssessmentAttempt{}, false, err
	}
	if !found {
		return AssessmentAttempt{}, false, nil
	}
	if err := decodeAttemptQuestionOrder(&row); err != nil {
		return AssessmentAttempt{}, false, err
	}
	return row, true, nil
}

func (model *AssessmentAttemptModel) ListByQuizAndUser(quizID uuid.UUID, userID string) ([]AssessmentAttempt, error) {
	var rows []AssessmentAttempt
	err := model.db.From(AssessmentAttemptsTable).
		Select(
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		).
		Where(goqu.Ex{"quiz_id": quizID, "user_id": userID}).
		Order(goqu.I("attempt_number").Asc()).
		ScanStructs(&rows)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []AssessmentAttempt{}
	}
	for i := range rows {
		if err := decodeAttemptQuestionOrder(&rows[i]); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (model *AssessmentAttemptModel) CountConsumingAttempts(quizID uuid.UUID, userID string) (int, error) {
	var count int64
	_, err := model.db.From(AssessmentAttemptsTable).
		Select(goqu.COUNT("*")).
		Where(goqu.Ex{
			"quiz_id": quizID,
			"user_id": userID,
		}).
		Where(goqu.C("status").Neq(AttemptStatusAbandoned)).
		ScanVal(&count)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (model *AssessmentAttemptModel) NextAttemptNumber(quizID uuid.UUID, userID string) (int, error) {
	var maxNumber sql.NullInt64
	_, err := model.db.From(AssessmentAttemptsTable).
		Select(goqu.MAX("attempt_number")).
		Where(goqu.Ex{"quiz_id": quizID, "user_id": userID}).
		ScanVal(&maxNumber)
	if err != nil {
		return 0, err
	}
	if !maxNumber.Valid {
		return 1, nil
	}
	return int(maxNumber.Int64) + 1, nil
}

// CreateInProgress inserts a new IN_PROGRESS attempt and copies resolved snapshot items
// in one transaction. On one-active conflict, returns the existing attempt with created=false.
func (model *AssessmentAttemptModel) CreateInProgress(params CreateAssessmentAttemptParams) (AssessmentAttempt, bool, error) {
	if params.TestSnapshotID == uuid.Nil {
		return AssessmentAttempt{}, false, ErrAssessmentAttemptSnapshotRequired
	}
	if len(params.QuestionOrder) == 0 {
		return AssessmentAttempt{}, false, ErrAssessmentAttemptEmptySnapshot
	}
	if len(params.SnapshotItems) == 0 {
		return AssessmentAttempt{}, false, ErrAttemptSnapshotItemsEmpty
	}
	if params.UserID == "" {
		return AssessmentAttempt{}, false, ErrAssessmentAttemptOwnerMismatch
	}

	attemptID, err := model.newUUID()
	if err != nil {
		return AssessmentAttempt{}, false, err
	}
	orderJSON, err := json.Marshal(params.QuestionOrder)
	if err != nil {
		return AssessmentAttempt{}, false, err
	}

	startedAt := params.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	now := time.Now().UTC()

	tx, err := model.db.Begin()
	if err != nil {
		return AssessmentAttempt{}, false, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = tx.Rollback()
		}
	}()

	record := goqu.Record{
		"id":                          attemptID,
		"quiz_id":                     params.QuizID,
		"user_id":                     params.UserID,
		"test_snapshot_id":            params.TestSnapshotID,
		"attempt_number":              params.AttemptNumber,
		"status":                      AttemptStatusInProgress,
		"question_order":              orderJSON,
		"negative_marks_per_question": params.NegativeMarksPerQuestion,
		"expected_max_score":          params.ExpectedMaxScore,
		"started_at":                  startedAt,
		"created_at":                  now,
		"updated_at":                  now,
	}
	if params.ExpiresAt != nil {
		record["expires_at"] = *params.ExpiresAt
	}

	_, err = tx.Insert(AssessmentAttemptsTable).Rows(record).Executor().Exec()
	if err != nil {
		if isOneActiveAttemptConflict(err) {
			_ = tx.Rollback()
			ok = true
			existing, found, getErr := model.GetInProgress(params.QuizID, params.UserID)
			if getErr != nil {
				return AssessmentAttempt{}, false, getErr
			}
			if found {
				return existing, false, nil
			}
			return AssessmentAttempt{}, false, err
		}
		return AssessmentAttempt{}, false, err
	}

	itemModel := InitAssessmentAttemptSnapshotItemModel(model.db)
	itemModel.newUUID = model.newUUID
	if err := itemModel.InsertItemsTx(tx, attemptID, params.SnapshotItems); err != nil {
		return AssessmentAttempt{}, false, err
	}

	if params.BeforeCommit != nil {
		if err := params.BeforeCommit(tx, attemptID); err != nil {
			return AssessmentAttempt{}, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return AssessmentAttempt{}, false, err
	}
	ok = true
	created, err := model.GetByID(params.QuizID, attemptID)
	return created, true, err
}

func isOneActiveAttemptConflict(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		return false
	}
	return pqErr.Constraint == assessmentAttemptsOneActiveConstraint ||
		pqErr.Constraint == "assessment_attempts_unique_per_attempt"
}

// GetByIDForUpdate loads an attempt row with FOR UPDATE inside a transaction.
func (model *AssessmentAttemptModel) GetByIDForUpdate(tx *goqu.TxDatabase, quizID, attemptID uuid.UUID) (AssessmentAttempt, error) {
	var row AssessmentAttempt
	found, err := tx.From(AssessmentAttemptsTable).
		Select(
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		).
		Where(goqu.Ex{"id": attemptID, "quiz_id": quizID}).
		ForUpdate(goqu.Wait).
		ScanStruct(&row)
	if err != nil {
		return AssessmentAttempt{}, err
	}
	if !found {
		return AssessmentAttempt{}, ErrAssessmentAttemptNotFound
	}
	if err := decodeAttemptQuestionOrder(&row); err != nil {
		return AssessmentAttempt{}, err
	}
	return row, nil
}

// TouchUpdatedAtTx bumps assessment_attempts.updated_at inside the answer transaction.
func (model *AssessmentAttemptModel) TouchUpdatedAtTx(tx *goqu.TxDatabase, quizID, attemptID uuid.UUID, now time.Time) error {
	_, err := tx.Update(AssessmentAttemptsTable).
		Set(goqu.Record{"updated_at": now}).
		Where(goqu.Ex{"id": attemptID, "quiz_id": quizID}).
		Executor().Exec()
	return err
}

// GetStatusByID returns lightweight status and timer fields for an attempt.
func (model *AssessmentAttemptModel) GetStatusByID(quizID, attemptID uuid.UUID) (AssessmentAttempt, error) {
	var row AssessmentAttempt
	found, err := model.db.From(AssessmentAttemptsTable).
		Select("id", "quiz_id", "user_id", "status", "expires_at").
		Where(goqu.Ex{"id": attemptID, "quiz_id": quizID}).
		ScanStruct(&row)
	if err != nil {
		return AssessmentAttempt{}, err
	}
	if !found {
		return AssessmentAttempt{}, ErrAssessmentAttemptNotFound
	}
	return row, nil
}

// FinalizeAttemptTx marks an attempt terminal (SUBMITTED or AUTO_SUBMITTED) with aggregate scores inside a transaction.
// Returns ErrAssessmentAttemptSubmitConflict when another writer already left IN_PROGRESS.
func (model *AssessmentAttemptModel) FinalizeAttemptTx(
	tx *goqu.TxDatabase,
	quizID, attemptID uuid.UUID,
	totalScore, maxScore float64,
	timeTakenSeconds int,
	submittedAt time.Time,
	status string,
) error {
	if status != AttemptStatusSubmitted && status != AttemptStatusAutoSubmitted {
		status = AttemptStatusSubmitted
	}
	result, err := tx.Update(AssessmentAttemptsTable).
		Set(goqu.Record{
			"status":             status,
			"submitted_at":       submittedAt,
			"total_score":        totalScore,
			"max_score":          maxScore,
			"time_taken_seconds": timeTakenSeconds,
			"updated_at":         submittedAt,
		}).
		Where(goqu.Ex{
			"id":      attemptID,
			"quiz_id": quizID,
			"status":  AttemptStatusInProgress,
		}).
		Executor().Exec()
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAssessmentAttemptSubmitConflict
	}
	return nil
}

func decodeAttemptQuestionOrder(attempt *AssessmentAttempt) error {
	if len(attempt.QuestionOrderJSON) == 0 {
		attempt.QuestionOrder = []uuid.UUID{}
		return nil
	}
	var order []uuid.UUID
	if err := json.Unmarshal(attempt.QuestionOrderJSON, &order); err != nil {
		return err
	}
	attempt.QuestionOrder = order
	return nil
}
