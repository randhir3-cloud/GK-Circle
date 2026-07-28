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
)

func newAssessmentAttemptModelTest(t *testing.T) (*AssessmentAttemptModel, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return InitAssessmentAttemptModel(goqu.New("postgres", sqlDB)), mock
}

func sampleFreezeItem(questionID uuid.UUID) CreateAttemptSnapshotItemParams {
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	points := int16(1)
	return CreateAttemptSnapshotItemParams{
		SnapshotItemID:     uuid.New(),
		Position:           0,
		QuestionID:         questionID,
		LineageID:          uuid.New(),
		RevisionNumber:     1,
		Question:           "Q?",
		Type:               1,
		OptionsJSON:        options,
		AnswersJSON:        answers,
		OfficialAnswerJSON: answers,
		AuthoritativeJSON:  answers,
		AnswerReviewStatus: "CONFIRMED",
		Points:             &points,
		QuestionMedia:      "text",
		OptionsMedia:       "text",
	}
}

func TestAssessmentAttemptCreateRequiresSnapshot(t *testing.T) {
	model, _ := newAssessmentAttemptModelTest(t)
	_, _, err := model.CreateInProgress(CreateAssessmentAttemptParams{
		QuizID:        uuid.New(),
		UserID:        "user-1",
		QuestionOrder: []uuid.UUID{uuid.New()},
	})
	if err != ErrAssessmentAttemptSnapshotRequired {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptCreateRequiresQuestionOrder(t *testing.T) {
	model, _ := newAssessmentAttemptModelTest(t)
	_, _, err := model.CreateInProgress(CreateAssessmentAttemptParams{
		QuizID:         uuid.New(),
		UserID:         "user-1",
		TestSnapshotID: uuid.New(),
	})
	if err != ErrAssessmentAttemptEmptySnapshot {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptCreatePersistsImmutableSnapshotBinding(t *testing.T) {
	model, mock := newAssessmentAttemptModelTest(t)
	quizID := uuid.New()
	snapshotID := uuid.MustParse("019c0200-0000-7000-8000-000000000001")
	attemptID := uuid.MustParse("019c0200-0000-7000-8000-000000000002")
	itemID := uuid.MustParse("019c0200-0000-7000-8000-000000000004")
	questionID := uuid.MustParse("019c0200-0000-7000-8000-000000000003")
	now := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})

	call := 0
	model.newUUID = func() (uuid.UUID, error) {
		call++
		if call == 1 {
			return attemptID, nil
		}
		return itemID, nil
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "assessment_attempt_snapshot_items"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "user-1", snapshotID, 1, AttemptStatusInProgress,
			orderJSON, 0.5, 1.0, now, nil, nil, nil, nil, nil, now, now,
		))

	attempt, created, err := model.CreateInProgress(CreateAssessmentAttemptParams{
		QuizID:                   quizID,
		UserID:                   "user-1",
		TestSnapshotID:           snapshotID,
		AttemptNumber:            1,
		QuestionOrder:            []uuid.UUID{questionID},
		NegativeMarksPerQuestion: 0.5,
		ExpectedMaxScore:         1,
		StartedAt:                now,
		SnapshotItems:            []CreateAttemptSnapshotItemParams{sampleFreezeItem(questionID)},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}
	if attempt.TestSnapshotID != snapshotID {
		t.Fatalf("snapshot binding = %s", attempt.TestSnapshotID)
	}
	if attempt.NegativeMarksPerQuestion != 0.5 {
		t.Fatalf("negative marks = %v", attempt.NegativeMarksPerQuestion)
	}
	if len(attempt.QuestionOrder) != 1 || attempt.QuestionOrder[0] != questionID {
		t.Fatalf("question_order = %+v", attempt.QuestionOrder)
	}
}

func TestAssessmentAttemptCreateOneActiveConflictReturnsExisting(t *testing.T) {
	model, mock := newAssessmentAttemptModelTest(t)
	quizID := uuid.New()
	snapshotID := uuid.New()
	existingID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})

	model.newUUID = func() (uuid.UUID, error) { return uuid.New(), nil }

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "assessment_attempts"`).
		WillReturnError(&pq.Error{Code: "23505", Constraint: assessmentAttemptsOneActiveConstraint})
	mock.ExpectRollback()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			existingID, quizID, "user-1", snapshotID, 1, AttemptStatusInProgress,
			orderJSON, 0.0, 1.0, now, nil, nil, nil, nil, nil, now, now,
		))

	attempt, created, err := model.CreateInProgress(CreateAssessmentAttemptParams{
		QuizID:         quizID,
		UserID:         "user-1",
		TestSnapshotID: snapshotID,
		AttemptNumber:  1,
		QuestionOrder:  []uuid.UUID{questionID},
		StartedAt:      now,
		SnapshotItems:  []CreateAttemptSnapshotItemParams{sampleFreezeItem(questionID)},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created {
		t.Fatal("expected created=false on conflict")
	}
	if attempt.ID != existingID {
		t.Fatalf("id = %s", attempt.ID)
	}
}

func TestAttemptAnswerModelListEmpty(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	model := InitAttemptAnswerModel(goqu.New("postgres", sqlDB))
	attemptID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))

	rows, err := model.ListByAttemptID(attemptID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("len = %d", len(rows))
	}
}
