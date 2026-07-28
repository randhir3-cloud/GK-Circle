package models

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func TestBuildImportPreviewRow_AppliesAuthorityDefaults(t *testing.T) {
	row, err := BuildImportPreviewRow(2, Question{
		Question: "Stem",
		Type:     1,
		Options:  map[string]string{"1": "A", "2": "B"},
		Answers:  []int{2},
		Points:   5,
		Resource: sql.NullString{String: "", Valid: true},
	})
	if err != nil {
		t.Fatalf("BuildImportPreviewRow: %v", err)
	}
	if row.Answers[0] != 2 {
		t.Fatalf("answers = %#v", row.Answers)
	}
	if row.OfficialAnswer[0] != 2 || row.AuthoritativeAnswer[0] != 2 {
		t.Fatalf("authority keys official=%#v authoritative=%#v", row.OfficialAnswer, row.AuthoritativeAnswer)
	}
	if row.AnswerReviewStatus != constants.AnswerReviewUnreviewed {
		t.Fatalf("status = %s", row.AnswerReviewStatus)
	}
	if row.RevisionNumber != 1 {
		t.Fatalf("revision = %d", row.RevisionNumber)
	}
}

func TestQuestionImportJobModel_CreateAndGet(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	model := InitQuestionImportJobModel(goqu.New("postgres", sqlDB), zap.NewNop())
	quizID := uuid.New()
	preview := ImportPreviewPayload{
		ValidRows: []ImportPreviewRow{{
			RowNumber:           2,
			Question:            "Stem",
			Type:                1,
			Options:             map[string]string{"1": "A", "2": "B"},
			Answers:             []int{1},
			OfficialAnswer:      []int{1},
			AuthoritativeAnswer: []int{1},
			AnswerReviewStatus:  constants.AnswerReviewUnreviewed,
			RevisionNumber:      1,
		}},
		Errors: []ImportRowError{{
			RowNumber: 3,
			Messages:  []string{constants.ErrEmptyQuestionText},
		}},
	}

	mock.ExpectExec(`INSERT INTO "question_import_jobs"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	job, err := model.CreatePreviewJob(quizID.String(), "user-1", "bank.csv", preview)
	if err != nil {
		t.Fatalf("CreatePreviewJob: %v", err)
	}
	if job.Status != ImportJobStatusPreviewed {
		t.Fatalf("status = %s", job.Status)
	}
	if job.ValidRowCount != 1 || job.ErrorRowCount != 1 || job.TotalRows != 2 {
		t.Fatalf("counts valid=%d error=%d total=%d", job.ValidRowCount, job.ErrorRowCount, job.TotalRows)
	}

	previewBytes, _ := json.Marshal(preview)
	commitBytes := []byte(`{}`)
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at",
			"created_at", "updated_at",
		}).AddRow(
			job.ID, quizID, "user-1", ImportJobStatusPreviewed, "bank.csv",
			2, 1, 1, previewBytes, commitBytes, nil, job.CreatedAt, job.UpdatedAt,
		))

	loaded, err := model.GetByQuizAndID(quizID.String(), job.ID.String())
	if err != nil {
		t.Fatalf("GetByQuizAndID: %v", err)
	}
	if len(loaded.Preview.ValidRows) != 1 || len(loaded.Preview.Errors) != 1 {
		t.Fatalf("loaded preview %#v", loaded.Preview)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestQuestionImportJobModel_CreateFailedWhenNoValidRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	model := InitQuestionImportJobModel(goqu.New("postgres", sqlDB), zap.NewNop())
	quizID := uuid.New()
	preview := ImportPreviewPayload{
		Errors: []ImportRowError{{RowNumber: 2, Messages: []string{constants.ErrEmptyQuestionText}}},
	}

	mock.ExpectExec(`INSERT INTO "question_import_jobs"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	job, err := model.CreatePreviewJob(quizID.String(), "user-1", "bad.csv", preview)
	if err != nil {
		t.Fatalf("CreatePreviewJob: %v", err)
	}
	if job.Status != ImportJobStatusFailed {
		t.Fatalf("status = %s, want FAILED", job.Status)
	}
}
