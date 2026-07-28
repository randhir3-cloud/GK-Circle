package services

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func newAttemptServiceTest(t *testing.T) (*AssessmentAttemptService, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db := goqu.New("postgres", sqlDB)
	return NewAssessmentAttemptService(db, zap.NewNop()), mock
}

func attemptSelectColumns() []string {
	return []string{
		"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
		"question_order", "negative_marks_per_question", "expected_max_score",
		"started_at", "submitted_at", "expires_at",
		"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
	}
}

func attemptSnapshotItemColumns() []string {
	return []string{
		"id", "attempt_id", "snapshot_item_id", "position", "question_id", "lineage_id", "revision_number",
		"question", "type", "options", "answers", "official_answer", "authoritative_answer",
		"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
	}
}

func expectSelfPacedQuiz(mock sqlmock.Sqlmock, quizID uuid.UUID, mode, status string, maxAttempts int, duration sql.NullInt64) {
	expectSelfPacedQuizWithNeg(mock, quizID, mode, status, maxAttempts, duration, 0.0)
}

func expectSelfPacedQuizWithNeg(mock sqlmock.Sqlmock, quizID uuid.UUID, mode, status string, maxAttempts int, duration sql.NullInt64, neg float64) {
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status", "duration_seconds", "max_attempts",
			"negative_marks_per_question", "allow_answer_review",
		}).AddRow(quizID, "PCS Practice", "Instructions body", "editor-1", true, mode, status, duration, maxAttempts, neg, false))
}

func expectQuizMetaForAnalytics(mock sqlmock.Sqlmock, quizID uuid.UUID) {
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(
			quizID, "PCS Practice", "Instructions body", "editor-1", true, "SELF_PACED", "PUBLISHED",
			int64(1800), 5, 0.0, true, "IMMEDIATE", true, nil, nil, true, true, true, true,
		))
}

func expectAnalyticsEventInsert(mock sqlmock.Sqlmock) {
	mock.ExpectExec(`INSERT INTO "assessment_analytics_events"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func expectServerTelemetryAfterCommit(mock sqlmock.Sqlmock, quizID uuid.UUID) {
	expectQuizMetaForAnalytics(mock, quizID)
	mock.ExpectBegin()
	expectAnalyticsEventInsert(mock)
	mock.ExpectCommit()
}

func expectNoInProgress(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()))
}

func expectSharedSnapshot(mock sqlmock.Sqlmock, snapshotID, quizID, questionID, lineageID uuid.UUID, sourceJSON, options, answers []byte, now time.Time, points int16) {
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at",
		}).AddRow(snapshotID, quizID, "editor-1", "CREATED", sourceJSON, 1, now))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "snapshot_id", "position", "collection_id", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), snapshotID, 0, nil, questionID, lineageID, 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			points, 30, "text", "text", nil, now,
		))
}

func expectAttemptSnapshotItemList(mock sqlmock.Sqlmock, attemptID, questionID uuid.UUID, options, answers []byte, now time.Time, points int16) {
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			points, 30, "text", "text", nil, now,
		))
}

func TestAssessmentAttemptServiceCreate_AuthorisedLearner(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	snapshotID := uuid.New()
	attemptID := uuid.New()
	itemRowID := uuid.New()
	questionID := uuid.New()
	lineageID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})

	call := 0
	svc.attemptModel.SetUUIDGenerator(func() (uuid.UUID, error) {
		call++
		if call == 1 {
			return attemptID, nil
		}
		return itemRowID, nil
	})

	expectSelfPacedQuizWithNeg(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 3, sql.NullInt64{}, 0.25)
	expectNoInProgress(mock)
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	expectSharedSnapshot(mock, snapshotID, quizID, questionID, lineageID, sourceJSON, options, answers, now, 5)
	mock.ExpectQuery(`SELECT MAX`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "assessment_attempt_snapshot_items"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	expectAnalyticsEventInsert(mock)
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.25, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
	expectSharedSnapshot(mock, snapshotID, quizID, questionID, lineageID, sourceJSON, options, answers, now, 5)
	expectAttemptSnapshotItemList(mock, attemptID, questionID, options, answers, now, 5)

	view, created, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: snapshotID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !created {
		t.Fatal("expected created")
	}
	if view.Attempt.TestSnapshotID != snapshotID {
		t.Fatalf("snapshot = %s", view.Attempt.TestSnapshotID)
	}
	if view.Attempt.NegativeMarksPerQuestion != 0.25 {
		t.Fatalf("frozen neg marks = %v", view.Attempt.NegativeMarksPerQuestion)
	}
	if !view.Attempt.ExpectedMaxScore.Valid || view.Attempt.ExpectedMaxScore.Float64 != 5 {
		t.Fatalf("expected max = %+v", view.Attempt.ExpectedMaxScore)
	}
	payload, _ := json.Marshal(view.Snapshot)
	if regexp.MustCompile(`"answers"|"official_answer"|"authoritative_answer"|"answer_review_status"`).Match(payload) {
		t.Fatalf("learner view leaked keys: %s", payload)
	}
}

func TestAssessmentAttemptServiceCreate_RejectsLiveQuiz(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	expectSelfPacedQuiz(mock, quizID, "LIVE", QuizStatusPublished, 1, sql.NullInt64{})

	_, _, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: uuid.New(),
	})
	if err != models.ErrAssessmentAttemptNotSelfPaced {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceCreate_RejectsUnpublishedWithoutEditor(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	expectSelfPacedQuiz(mock, quizID, QuizAssessmentModeSelfPaced, "DRAFT", 1, sql.NullInt64{})

	_, _, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: uuid.New(),
	})
	if err != models.ErrAssessmentAttemptQuizNotPublished {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceCreate_IdempotentInProgress(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	snapshotID := uuid.New()
	attemptID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})

	expectSelfPacedQuiz(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 3, sql.NullInt64{})
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
	expectSharedSnapshot(mock, snapshotID, quizID, questionID, uuid.New(), sourceJSON, options, answers, now, 5)
	expectAttemptSnapshotItemList(mock, attemptID, questionID, options, answers, now, 5)

	view, created, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: uuid.New(), // ignored when in-progress exists
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created {
		t.Fatal("expected idempotent reuse")
	}
	if view.Attempt.ID != attemptID {
		t.Fatalf("id = %s", view.Attempt.ID)
	}
}

func TestAssessmentAttemptServiceCreate_ForeignSnapshot(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	expectSelfPacedQuiz(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 1, sql.NullInt64{})
	expectNoInProgress(mock)
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at",
		}))

	_, _, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: uuid.New(),
	})
	if err != models.ErrAssessmentAttemptForeignSnapshot {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceGetLearner_Ownership(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{})

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "owner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 0.0, now, nil, nil, nil, nil, nil, now, now,
		))

	_, err := svc.GetLearner(quizID, attemptID, "other-user")
	if err != models.ErrAssessmentAttemptNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceCreate_EmptySnapshot(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	snapshotID := uuid.New()
	now := time.Now().UTC()
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})

	expectSelfPacedQuiz(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 1, sql.NullInt64{})
	expectNoInProgress(mock)
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at",
		}).AddRow(snapshotID, quizID, "editor-1", "CREATED", sourceJSON, 0, now))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "snapshot_id", "position", "collection_id", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}))

	_, _, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: snapshotID,
	})
	if err != models.ErrAssessmentAttemptEmptySnapshot {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceCreate_MaxAttempts(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	expectSelfPacedQuiz(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 1, sql.NullInt64{})
	expectNoInProgress(mock)
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, _, err := svc.Create(CreateAssessmentAttemptRequest{
		QuizID:     quizID,
		UserID:     "learner-1",
		SnapshotID: uuid.New(),
	})
	if err != models.ErrAssessmentAttemptMaxReached {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceGetInstructions_StartEligible(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	duration := sql.NullInt64{Int64: 3600, Valid: true}

	expectSelfPacedQuizWithNeg(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 3, duration, 0.25)
	expectSharedSnapshot(mock, snapshotID, quizID, questionID, uuid.New(), sourceJSON, options, answers, now, 5)
	expectNoInProgress(mock)
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	view, err := svc.GetInstructions(quizID, snapshotID, "learner-1", false)
	if err != nil {
		t.Fatalf("instructions: %v", err)
	}
	if !view.CanStart || view.CanResume {
		t.Fatalf("eligibility = can_start=%v can_resume=%v", view.CanStart, view.CanResume)
	}
	if view.QuestionCount != 1 || view.Quiz.Title != "PCS Practice" {
		t.Fatalf("view = %+v", view)
	}
	if view.Quiz.NegativeMarksPerQuestion != 0.25 {
		t.Fatalf("neg marks = %v", view.Quiz.NegativeMarksPerQuestion)
	}
}

func TestAssessmentAttemptServiceGetInstructions_ResumePreferred(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	snapshotID := uuid.New()
	attemptID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})

	expectSelfPacedQuiz(mock, quizID, QuizAssessmentModeSelfPaced, QuizStatusPublished, 3, sql.NullInt64{})
	expectSharedSnapshot(mock, snapshotID, quizID, questionID, uuid.New(), sourceJSON, options, answers, now, 5)
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))

	view, err := svc.GetInstructions(quizID, snapshotID, "learner-1", false)
	if err != nil {
		t.Fatalf("instructions: %v", err)
	}
	if view.CanStart || !view.CanResume || view.ActiveAttempt == nil || view.ActiveAttempt.ID != attemptID {
		t.Fatalf("view = %+v", view)
	}
}
