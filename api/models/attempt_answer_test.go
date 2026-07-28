package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func TestValidateSelectedOptionsAgainstSnapshot_SingleAndSurvey(t *testing.T) {
	options := map[string]string{"1": "A", "2": "B", "3": "C"}

	if err := ValidateSelectedOptionsAgainstSnapshot(constants.SingleAnswer, options, []int{1}, false); err != nil {
		t.Fatalf("single ok: %v", err)
	}
	if err := ValidateSelectedOptionsAgainstSnapshot(constants.SingleAnswer, options, []int{1, 2}, false); err != ErrAttemptAnswerCardinality {
		t.Fatalf("single multi = %v", err)
	}
	if err := ValidateSelectedOptionsAgainstSnapshot(constants.SingleAnswer, options, []int{9}, false); err != ErrAttemptAnswerInvalidOptionRef {
		t.Fatalf("foreign option = %v", err)
	}
	if err := ValidateSelectedOptionsAgainstSnapshot(constants.Survey, options, []int{1, 3}, false); err != nil {
		t.Fatalf("survey ok: %v", err)
	}
	if err := ValidateSelectedOptionsAgainstSnapshot(constants.Survey, options, []int{1, 1}, false); err != ErrAttemptAnswerInvalidOptions {
		t.Fatalf("dup = %v", err)
	}
	if err := ValidateSelectedOptionsAgainstSnapshot(constants.SingleAnswer, options, nil, true); err != nil {
		t.Fatalf("clear: %v", err)
	}
}

func TestAttemptAnswerUpsertInsertAndUpdate(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db := goqu.New("postgres", sqlDB)
	model := InitAttemptAnswerModel(db)

	attemptID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.MustParse("019c0300-0000-7000-8000-000000000001")
	now := time.Now().UTC()
	model.newUUID = func() (uuid.UUID, error) { return answerID, nil }

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))
	mock.ExpectExec(`INSERT INTO "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	selected, _ := json.Marshal([]int{1})
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selected, false, now, nil, nil, nil, now, now))

	created, err := model.UpsertAnswer(tx, UpsertAttemptAnswerParams{
		AttemptID:       attemptID,
		QuestionID:      questionID,
		SelectedOptions: []int{1},
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if created.ID != answerID {
		t.Fatalf("id = %s", created.ID)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selected, false, now, nil, nil, nil, now, now))
	mock.ExpectExec(`UPDATE "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	updatedSelected, _ := json.Marshal([]int{2})
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, updatedSelected, true, now, 5, nil, nil, now, now))

	updated, err := model.UpsertAnswer(tx, UpsertAttemptAnswerParams{
		AttemptID:        attemptID,
		QuestionID:       questionID,
		SelectedOptions:  []int{2},
		IsMarkedReview:   true,
		TimeTakenSeconds: intPtr(5),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	opts, _ := DecodeSelectedOptions(updated.SelectedOptions)
	if len(opts) != 1 || opts[0] != 2 {
		t.Fatalf("options = %+v", opts)
	}
	if !updated.IsMarkedReview {
		t.Fatal("expected marked review")
	}
	_ = tx.Rollback()
}

func TestAttemptAnswerUpsertConcurrentInsertFallsBackToUpdate(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db := goqu.New("postgres", sqlDB)
	model := InitAttemptAnswerModel(db)

	attemptID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.New()
	now := time.Now().UTC()
	selected, _ := json.Marshal([]int{1})
	model.newUUID = func() (uuid.UUID, error) { return uuid.New(), nil }

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))
	mock.ExpectExec(`INSERT INTO "attempt_answers"`).
		WillReturnError(&pq.Error{Code: "23505", Constraint: attemptAnswersUniqueConstraint})
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selected, false, now, nil, nil, nil, now, now))
	mock.ExpectExec(`UPDATE "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selected, false, now, nil, nil, nil, now, now))

	answer, err := model.UpsertAnswer(tx, UpsertAttemptAnswerParams{
		AttemptID:       attemptID,
		QuestionID:      questionID,
		SelectedOptions: []int{1},
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if answer.ID != answerID {
		t.Fatalf("id = %s", answer.ID)
	}
	_ = tx.Rollback()
}

func intPtr(v int) *int { return &v }
