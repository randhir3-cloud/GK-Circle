package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func TestQuestionsFromImportPreviewRows_ReappliesAuthority(t *testing.T) {
	rows := []ImportPreviewRow{{
		Question:            "Stem",
		Type:                1,
		Options:             map[string]string{"1": "A", "2": "B"},
		Answers:             []int{2},
		OfficialAnswer:      nil,
		AuthoritativeAnswer: nil,
		AnswerReviewStatus:  constants.AnswerReviewUnreviewed,
		Points:              5,
		RevisionNumber:      1,
	}}

	questions, err := QuestionsFromImportPreviewRows(rows)
	if err != nil {
		t.Fatalf("QuestionsFromImportPreviewRows: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("questions = %d", len(questions))
	}
	if questions[0].OfficialAnswer[0] != 2 || questions[0].AuthoritativeAnswer[0] != 2 {
		t.Fatalf("authority not revalidated: official=%#v authoritative=%#v",
			questions[0].OfficialAnswer, questions[0].AuthoritativeAnswer)
	}
	if questions[0].RevisionNumber != 1 {
		t.Fatalf("revision = %d", questions[0].RevisionNumber)
	}
}

func TestQuestionImportJobModel_TryClaimForCommit(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	model := InitQuestionImportJobModel(goqu.New("postgres", sqlDB), zap.NewNop())
	quizID := uuid.New()
	jobID := uuid.New()
	preview := ImportPreviewPayload{
		ValidRows: []ImportPreviewRow{{
			RowNumber: 2, Question: "Stem", Type: 1,
			Options: map[string]string{"1": "A", "2": "B"}, Answers: []int{1},
			OfficialAnswer: []int{1}, AuthoritativeAnswer: []int{1},
			AnswerReviewStatus: constants.AnswerReviewUnreviewed, RevisionNumber: 1,
		}},
	}
	previewBytes, _ := json.Marshal(preview)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "user-1", ImportJobStatusPreviewed, "bank.csv",
			1, 1, 0, previewBytes, []byte(`{}`), nil, time.Now(), time.Now(),
		))
	mock.ExpectExec(`UPDATE "question_import_jobs"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tx, err := goqu.New("postgres", sqlDB).Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	job, claimed, err := model.TryClaimForCommit(tx, quizID.String(), jobID.String())
	if err != nil {
		t.Fatalf("TryClaimForCommit: %v", err)
	}
	if !claimed {
		t.Fatal("expected claim")
	}
	if job.Status != ImportJobStatusCommitting {
		t.Fatalf("status = %s", job.Status)
	}
	if len(job.Preview.ValidRows) != 1 {
		t.Fatalf("preview rows = %d", len(job.Preview.ValidRows))
	}
}

func TestQuestionImportJobModel_FinalizeCommit(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	model := InitQuestionImportJobModel(goqu.New("postgres", sqlDB), zap.NewNop())
	quizID := uuid.New()
	jobID := uuid.New()
	committedAt := time.Now()
	result := ImportCommitResult{QuestionIDs: []string{uuid.New().String()}, CommittedCount: 1}
	resultBytes, _ := json.Marshal(result)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "user-1", ImportJobStatusCommitted, "bank.csv",
			1, 1, 0, []byte(`{"valid_rows":[],"errors":[]}`), resultBytes, committedAt, committedAt, committedAt,
		))

	tx, err := goqu.New("postgres", sqlDB).Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	job, err := model.FinalizeCommit(tx, quizID.String(), jobID.String(), result)
	if err != nil {
		t.Fatalf("FinalizeCommit: %v", err)
	}
	if job.Status != ImportJobStatusCommitted {
		t.Fatalf("status = %s", job.Status)
	}
	if job.CommitResult.CommittedCount != 1 {
		t.Fatalf("committed count = %d", job.CommitResult.CommittedCount)
	}
}

func TestQuestionImportJobModel_MarkCommitFailed(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	model := InitQuestionImportJobModel(goqu.New("postgres", sqlDB), zap.NewNop())
	quizID := uuid.New()
	jobID := uuid.New()

	mock.ExpectExec(`UPDATE "question_import_jobs"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := model.MarkCommitFailed(quizID.String(), jobID.String(), "insert failed"); err != nil {
		t.Fatalf("MarkCommitFailed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
