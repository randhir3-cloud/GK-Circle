package services

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func TestAssessmentAttemptServiceSubmit_ScoresAndIdempotent(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	qCorrect := uuid.New()
	qWrong := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{qCorrect, qWrong})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A", "2": "B"})
	answers, _ := json.Marshal([]int{1})
	selectedCorrect, _ := json.Marshal([]int{1})
	selectedWrong, _ := json.Marshal([]int{2})

	answerID1 := uuid.New()
	answerID2 := uuid.New()
	call := 0
	svc.answerModel.SetUUIDGenerator(func() (uuid.UUID, error) {
		call++
		if call == 1 {
			return answerID1, nil
		}
		return answerID2, nil
	})

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.5, 2.0, now.Add(-2*time.Minute), nil, nil, nil, nil, nil, now, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).
			AddRow(
				uuid.New(), attemptID, uuid.New(), 0, qCorrect, uuid.New(), 1,
				"Q1", 1, options, answers, answers, answers, "CONFIRMED",
				int16(1), 30, "text", "text", nil, now,
			).AddRow(
				uuid.New(), attemptID, uuid.New(), 1, qWrong, uuid.New(), 1,
				"Q2", 1, options, answers, answers, answers, "CONFIRMED",
				int16(1), 30, "text", "text", nil, now,
			))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(uuid.New(), attemptID, qCorrect, selectedCorrect, false, now, nil, nil, nil, now, now).
			AddRow(uuid.New(), attemptID, qWrong, selectedWrong, false, now, nil, nil, nil, now, now))

	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID1, attemptID, qCorrect, selectedCorrect, false, now, nil, nil, nil, now, now))
	mock.ExpectExec(`UPDATE "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID2, attemptID, qWrong, selectedWrong, false, now, nil, nil, nil, now, now))
	mock.ExpectExec(`UPDATE "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	expectQuizMetaForAnalytics(mock, quizID)
	expectAnalyticsEventInsert(mock)
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusSubmitted,
			orderJSON, 0.5, 2.0, now.Add(-2*time.Minute), now, nil, 0.5, 2.0, 120, now, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at",
		}).AddRow(snapshotID, quizID, "editor-1", "CREATED", sourceJSON, 2, now))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "snapshot_id", "position", "collection_id", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), snapshotID, 0, nil, qCorrect, uuid.New(), 1,
			"Q1", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		).AddRow(
			uuid.New(), snapshotID, 1, nil, qWrong, uuid.New(), 1,
			"Q2", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))
	// learnerSnapshotFromAttempt + ordered answers both list attempt items
	for i := 0; i < 2; i++ {
		mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
			WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).
				AddRow(
					uuid.New(), attemptID, uuid.New(), 0, qCorrect, uuid.New(), 1,
					"Q1", 1, options, answers, answers, answers, "CONFIRMED",
					int16(1), 30, "text", "text", nil, now,
				).AddRow(
					uuid.New(), attemptID, uuid.New(), 1, qWrong, uuid.New(), 1,
					"Q2", 1, options, answers, answers, answers, "CONFIRMED",
					int16(1), 30, "text", "text", nil, now,
				))
	}
	trueVal := true
	falseVal := false
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID1, attemptID, qCorrect, selectedCorrect, false, now, nil, 1.0, trueVal, now, now).
			AddRow(answerID2, attemptID, qWrong, selectedWrong, false, now, nil, -0.5, falseVal, now, now))

	view, created, err := svc.Submit(quizID, attemptID, "learner-1", "corr-test")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if !created {
		t.Fatal("expected created")
	}
	if view.Summary.TotalScore != 0.5 || view.Summary.MaxScore != 2 {
		t.Fatalf("summary = %+v", view.Summary)
	}
	body := string(mustJSON(view))
	if regexp.MustCompile(`"official_answer"|"authoritative_answer"|"answer_review_status"`).MatchString(body) {
		t.Fatalf("result leaked keys: %s", body)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusSubmitted,
			orderJSON, 0.5, 2.0, now.Add(-2*time.Minute), now, nil, 0.5, 2.0, 120, now, now,
		))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshots"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_collection_ids", "question_count", "created_at",
		}).AddRow(snapshotID, quizID, "editor-1", "CREATED", sourceJSON, 2, now))
	mock.ExpectQuery(`SELECT (.+) FROM "test_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "snapshot_id", "position", "collection_id", "question_id", "lineage_id", "revision_number",
			"question", "type", "options", "answers", "official_answer", "authoritative_answer",
			"answer_review_status", "points", "duration_in_seconds", "question_media", "options_media", "resource", "created_at",
		}).AddRow(
			uuid.New(), snapshotID, 0, nil, qCorrect, uuid.New(), 1,
			"Q1", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		).AddRow(
			uuid.New(), snapshotID, 1, nil, qWrong, uuid.New(), 1,
			"Q2", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))
	for i := 0; i < 2; i++ {
		mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
			WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).
				AddRow(
					uuid.New(), attemptID, uuid.New(), 0, qCorrect, uuid.New(), 1,
					"Q1", 1, options, answers, answers, answers, "CONFIRMED",
					int16(1), 30, "text", "text", nil, now,
				).AddRow(
					uuid.New(), attemptID, uuid.New(), 1, qWrong, uuid.New(), 1,
					"Q2", 1, options, answers, answers, answers, "CONFIRMED",
					int16(1), 30, "text", "text", nil, now,
				))
	}
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID1, attemptID, qCorrect, selectedCorrect, false, now, nil, 1.0, trueVal, now, now).
			AddRow(answerID2, attemptID, qWrong, selectedWrong, false, now, nil, -0.5, falseVal, now, now))

	view2, created2, err := svc.Submit(quizID, attemptID, "learner-1", "corr-test")
	if err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	if created2 {
		t.Fatal("expected idempotent")
	}
	if view2.Summary.TotalScore != 0.5 {
		t.Fatalf("resubmit total = %v", view2.Summary.TotalScore)
	}
}

func TestAssessmentAttemptServiceSubmit_UsesFrozenNegMarksNotQuiz(t *testing.T) {
	// Submit must score from attempt-linked items + frozen attempt.negative_marks_per_question
	// without re-reading quizzes.negative_marks_per_question.
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A", "2": "B"})
	answers, _ := json.Marshal([]int{1})
	selectedWrong, _ := json.Marshal([]int{2})
	answerID := uuid.New()
	svc.answerModel.SetUUIDGenerator(func() (uuid.UUID, error) { return answerID, nil })

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.25, 1.0, now.Add(-time.Minute), nil, nil, nil, nil, nil, now, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selectedWrong, false, now, nil, nil, nil, now, now))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selectedWrong, false, now, nil, nil, nil, now, now))
	mock.ExpectExec(`UPDATE "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	expectQuizMetaForAnalytics(mock, quizID)
	expectAnalyticsEventInsert(mock)
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusSubmitted,
			orderJSON, 0.25, 1.0, now.Add(-time.Minute), now, nil, 0.0, 1.0, 60, now, now,
		))
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
			uuid.New(), snapshotID, 0, nil, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))
	for i := 0; i < 2; i++ {
		expectAttemptSnapshotItemList(mock, attemptID, questionID, options, answers, now, 1)
	}
	falseVal := false
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selectedWrong, false, now, nil, -0.25, falseVal, now, now))

	view, created, err := svc.Submit(quizID, attemptID, "learner-1", "corr-test")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if !created {
		t.Fatal("expected created")
	}
	// Incorrect answer with frozen 0.25 penalty floors to 0.
	if view.Summary.TotalScore != 0 || view.Summary.MaxScore != 1 {
		t.Fatalf("summary = %+v", view.Summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected quiz re-read or leftover expectations: %v", err)
	}
}

func TestAssessmentAttemptServiceSubmit_RejectsNonOwner(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{})

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "owner-1", uuid.New(), 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 0.0, now, nil, nil, nil, nil, nil, now, now,
		))
	mock.ExpectRollback()

	_, _, err := svc.Submit(quizID, attemptID, "intruder", "corr-test")
	if err != models.ErrAssessmentAttemptNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceSubmit_AutoSubmitOnExpiry(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	now := time.Now().UTC()
	pastExpiry := now.Add(-10 * time.Second)
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A", "2": "B"})
	answers, _ := json.Marshal([]int{1})
	answerID := uuid.New()
	svc.answerModel.SetUUIDGenerator(func() (uuid.UUID, error) { return answerID, nil })

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 1.0, now.Add(-5*time.Minute), nil, pastExpiry, nil, nil, nil, now, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))
	mock.ExpectExec(`(INSERT INTO|UPDATE) "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	expectQuizMetaForAnalytics(mock, quizID)
	expectAnalyticsEventInsert(mock)
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusAutoSubmitted,
			orderJSON, 0.0, 1.0, now.Add(-5*time.Minute), now, pastExpiry, 0.0, 1.0, 300, now, now,
		))
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
			uuid.New(), snapshotID, 0, nil, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(1), 30, "text", "text", nil, now,
		))
	for i := 0; i < 2; i++ {
		expectAttemptSnapshotItemList(mock, attemptID, questionID, options, answers, now, 1)
	}
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))

	view, created, err := svc.Submit(quizID, attemptID, "learner-1", "corr-test")
	if err != nil {
		t.Fatalf("submit error: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}
	if view.Attempt.Status != models.AttemptStatusAutoSubmitted {
		t.Fatalf("expected status %s, got %s", models.AttemptStatusAutoSubmitted, view.Attempt.Status)
	}
}

func TestAssessmentAttemptServiceGetStatus(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()
	expiry := now.Add(5 * time.Minute)

	mock.ExpectQuery(`SELECT "id", "quiz_id", "user_id", "status", "expires_at" FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "quiz_id", "user_id", "status", "expires_at"}).
			AddRow(attemptID, quizID, "learner-1", models.AttemptStatusInProgress, expiry))

	statusView, err := svc.GetStatus(quizID, attemptID, "learner-1")
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if statusView.Status != models.AttemptStatusInProgress {
		t.Fatalf("expected status IN_PROGRESS, got %s", statusView.Status)
	}
	if statusView.RemainingSeconds == nil || *statusView.RemainingSeconds <= 0 {
		t.Fatalf("expected positive remaining_seconds, got %v", statusView.RemainingSeconds)
	}
}
