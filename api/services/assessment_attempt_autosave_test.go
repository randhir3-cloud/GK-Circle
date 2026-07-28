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

func TestAssessmentAttemptServiceAutosave_RejectsForeignQuestion(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	foreignID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()))
	mock.ExpectRollback()

	_, err := svc.AutosaveAnswer(AutosaveAnswerRequest{
		QuizID:          quizID,
		AttemptID:       attemptID,
		QuestionID:      foreignID,
		UserID:          "learner-1",
		SelectedOptions: []int{1},
	})
	if err != models.ErrAttemptAnswerQuestionNotInSnapshot {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceAutosave_RejectsTerminalAttempt(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{})

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", uuid.New(), 1, models.AttemptStatusSubmitted,
			orderJSON, 0.0, 0.0, now, now, nil, nil, nil, nil, now, now,
		))
	mock.ExpectRollback()

	_, err := svc.AutosaveAnswer(AutosaveAnswerRequest{
		QuizID:          quizID,
		AttemptID:       attemptID,
		QuestionID:      uuid.New(),
		UserID:          "learner-1",
		SelectedOptions: []int{1},
	})
	if err != models.ErrAttemptAnswerNotInProgress {
		t.Fatalf("err = %v", err)
	}
}

func TestAssessmentAttemptServiceAutosave_AndResumeNoKeys(t *testing.T) {
	svc, mock := newAttemptServiceTest(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.New()
	now := time.Now().UTC()
	orderJSON, _ := json.Marshal([]uuid.UUID{questionID})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})
	options, _ := json.Marshal(map[string]string{"1": "A", "2": "B"})
	answers, _ := json.Marshal([]int{1})
	selected, _ := json.Marshal([]int{1})

	svc.answerModel.SetUUIDGenerator(func() (uuid.UUID, error) { return answerID, nil })

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempt_snapshot_items"`).
		WillReturnRows(sqlmock.NewRows(attemptSnapshotItemColumns()).AddRow(
			uuid.New(), attemptID, uuid.New(), 0, questionID, uuid.New(), 1,
			"Q?", 1, options, answers, answers, answers, "CONFIRMED",
			int16(5), 30, "text", "text", nil, now,
		))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}))
	mock.ExpectExec(`INSERT INTO "attempt_answers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selected, false, now, nil, nil, nil, now, now))
	mock.ExpectExec(`UPDATE "assessment_attempts"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	expectServerTelemetryAfterCommit(mock, quizID)

	saved, err := svc.AutosaveAnswer(AutosaveAnswerRequest{
		QuizID:          quizID,
		AttemptID:       attemptID,
		QuestionID:      questionID,
		UserID:          "learner-1",
		SelectedOptions: []int{1},
	})
	if err != nil {
		t.Fatalf("autosave: %v", err)
	}
	if len(saved.SelectedOptions) != 1 || saved.SelectedOptions[0] != 1 {
		t.Fatalf("selected = %+v", saved.SelectedOptions)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(sqlmock.NewRows(attemptSelectColumns()).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, models.AttemptStatusInProgress,
			orderJSON, 0.0, 5.0, now, nil, nil, nil, nil, nil, now, now,
		))
	expectSharedSnapshot(mock, snapshotID, quizID, questionID, uuid.New(), sourceJSON, options, answers, now, 5)
	expectAttemptSnapshotItemList(mock, attemptID, questionID, options, answers, now, 5)
	mock.ExpectQuery(`SELECT (.+) FROM "attempt_answers"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "attempt_id", "question_id", "selected_options", "is_marked_review",
			"answered_at", "time_taken_seconds", "score", "is_correct", "created_at", "updated_at",
		}).AddRow(answerID, attemptID, questionID, selected, false, now, nil, nil, nil, now, now))

	resume, err := svc.Resume(quizID, attemptID, "learner-1")
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	body := string(mustJSON(resume))
	if regexp.MustCompile(`"official_answer"|"authoritative_answer"|"answer_review_status"`).MatchString(body) {
		t.Fatalf("resume leaked keys: %s", body)
	}
	if regexp.MustCompile(`"is_correct"|"score"`).MatchString(body) {
		t.Fatalf("resume leaked score fields: %s", body)
	}
	if resume.Progress.AnsweredCount != 1 || resume.Progress.TotalQuestions != 1 {
		t.Fatalf("progress = %+v", resume.Progress)
	}
}

func TestAssessmentAttemptServiceAutosave_Ownership(t *testing.T) {
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

	_, err := svc.AutosaveAnswer(AutosaveAnswerRequest{
		QuizID:          quizID,
		AttemptID:       attemptID,
		QuestionID:      uuid.New(),
		UserID:          "intruder",
		SelectedOptions: []int{1},
	})
	if err != models.ErrAssessmentAttemptNotFound {
		t.Fatalf("err = %v", err)
	}
}

func mustJSON(v any) []byte {
	payload, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return payload
}
