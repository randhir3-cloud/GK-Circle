package v1

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "import.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	return path
}

func TestCreateQuestionImportJob_ReturnsPreviewWithoutPersistingQuestions(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	cfg.Quiz.QuestionTimeLimit = "45"
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	csvPath := writeTempCSV(t, `Question Text,Question Type,Points,Option 1,Option 2,Option 3,Option 4,Option 5,Correct Answer,Question Media,Options Media,Resource
Capital?,single answer,5,Paris,Berlin,,,,1,text,text,
,single answer,5,A,B,,,,1,text,text,
`)

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "category_id", "cover_image", "created_at", "updated_at",
		}).AddRow(quizID, "Demo", sql.NullString{}, "creator-1", false, sql.NullString{}, sql.NullString{}, time.Now(), time.Now()))

	mock.ExpectQuery(`SELECT (.+) FROM "quiz_questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"question_id", "question", "type", "options", "answers"}))

	mock.ExpectExec(`INSERT INTO "question_import_jobs"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/import-jobs", func(c *fiber.Ctx) error {
		c.Locals(constants.FileName, csvPath)
		c.Locals(constants.ContextUid, "editor-1")
		return c.Next()
	}, ctrl.CreateQuestionImportJob)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/import-jobs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var envelope struct {
		Data models.QuestionImportJob `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, body)
	}
	if envelope.Data.ValidRowCount != 1 || envelope.Data.ErrorRowCount != 1 {
		t.Fatalf("counts valid=%d error=%d", envelope.Data.ValidRowCount, envelope.Data.ErrorRowCount)
	}
	if len(envelope.Data.Preview.ValidRows) != 1 {
		t.Fatalf("preview valid rows = %d", len(envelope.Data.Preview.ValidRows))
	}
	if envelope.Data.Preview.ValidRows[0].AnswerReviewStatus != constants.AnswerReviewUnreviewed {
		t.Fatalf("authority status = %s", envelope.Data.Preview.ValidRows[0].AnswerReviewStatus)
	}
	if _, err := os.Stat(csvPath); !os.IsNotExist(err) {
		t.Fatalf("temp csv should be removed, err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGetQuestionImportJob_NotFound(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	cfg.Quiz.QuestionTimeLimit = "30"
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	jobID := uuid.New()
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at",
			"created_at", "updated_at",
		}))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/questions/import-jobs/:import_job_id", ctrl.GetQuestionImportJob)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/questions/import-jobs/"+jobID.String(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestCommitQuestionImportJob_AlreadyCommittedIsIdempotent(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	cfg.Quiz.QuestionTimeLimit = "30"
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	jobID := uuid.New()
	commitResult, _ := json.Marshal(models.ImportCommitResult{
		QuestionIDs:    []string{uuid.New().String()},
		CommittedCount: 1,
	})

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "category_id", "cover_image", "created_at", "updated_at",
		}).AddRow(quizID, "Demo", sql.NullString{}, "creator-1", false, sql.NullString{}, sql.NullString{}, time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "editor-1", models.ImportJobStatusCommitted, "bank.csv",
			1, 1, 0, []byte(`{"valid_rows":[],"errors":[]}`), commitResult, time.Now(), time.Now(), time.Now(),
		))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit", ctrl.CommitQuestionImportJob)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/import-jobs/"+jobID.String()+"/commit", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}
}

func TestCommitQuestionImportJob_InProgressConflict(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	jobID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "category_id", "cover_image", "created_at", "updated_at",
		}).AddRow(quizID, "Demo", sql.NullString{}, "creator-1", false, sql.NullString{}, sql.NullString{}, time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "editor-1", models.ImportJobStatusCommitting, "bank.csv",
			1, 1, 0, []byte(`{"valid_rows":[],"errors":[]}`), []byte(`{}`), nil, time.Now(), time.Now(),
		))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit", ctrl.CommitQuestionImportJob)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/import-jobs/"+jobID.String()+"/commit", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestCommitQuestionImportJob_NoValidRowsBadRequest(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	jobID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "category_id", "cover_image", "created_at", "updated_at",
		}).AddRow(quizID, "Demo", sql.NullString{}, "creator-1", false, sql.NullString{}, sql.NullString{}, time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "editor-1", models.ImportJobStatusPreviewed, "bank.csv",
			1, 0, 1, []byte(`{"valid_rows":[],"errors":[{"row_number":2,"messages":["empty"]}]}`), []byte(`{}`), nil, time.Now(), time.Now(),
		))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit", ctrl.CommitQuestionImportJob)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/import-jobs/"+jobID.String()+"/commit", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestCreateQuestionImportJob_FlagsDuplicateRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	cfg.Quiz.QuestionTimeLimit = "45"
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	csvPath := writeTempCSV(t, `Question Text,Question Type,Points,Option 1,Option 2,Option 3,Option 4,Option 5,Correct Answer,Question Media,Options Media,Resource
Capital?,single answer,5,Paris,Berlin,,,,1,text,text,
Capital?,single answer,5,Paris,Berlin,,,,1,text,text,
`)

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "category_id", "cover_image", "created_at", "updated_at",
		}).AddRow(quizID, "Demo", sql.NullString{}, "creator-1", false, sql.NullString{}, sql.NullString{}, time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT (.+) FROM "quiz_questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"question_id", "question", "type", "options", "answers"}))
	mock.ExpectExec(`INSERT INTO "question_import_jobs"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/import-jobs", func(c *fiber.Ctx) error {
		c.Locals(constants.FileName, csvPath)
		c.Locals(constants.ContextUid, "editor-1")
		return c.Next()
	}, ctrl.CreateQuestionImportJob)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/import-jobs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var envelope struct {
		Data models.QuestionImportJob `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.ValidRowCount != 1 || envelope.Data.ErrorRowCount != 1 {
		t.Fatalf("counts valid=%d error=%d", envelope.Data.ValidRowCount, envelope.Data.ErrorRowCount)
	}
	if len(envelope.Data.Preview.Errors) != 1 {
		t.Fatalf("errors = %d", len(envelope.Data.Preview.Errors))
	}
	if envelope.Data.Preview.Errors[0].Kind != models.ImportRowErrorKindDuplicate {
		t.Fatalf("kind = %s", envelope.Data.Preview.Errors[0].Kind)
	}
}

func TestCommitQuestionImportJob_BlocksQuizDuplicate(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := &config.AppConfig{}
	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), cfg)
	if err != nil {
		t.Fatalf("InitQuestionController: %v", err)
	}

	quizID := uuid.New()
	jobID := uuid.New()
	existingQuestionID := uuid.New()
	preview := models.ImportPreviewPayload{
		ValidRows: []models.ImportPreviewRow{{
			RowNumber: 2,
			Question:  "Capital?",
			Type:      1,
			Options:   map[string]string{"1": "Paris", "2": "Berlin"},
			Answers:   []int{1},
		}},
	}
	previewBytes, _ := json.Marshal(preview)

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "category_id", "cover_image", "created_at", "updated_at",
		}).AddRow(quizID, "Demo", sql.NullString{}, "creator-1", false, sql.NullString{}, sql.NullString{}, time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "editor-1", models.ImportJobStatusPreviewed, "bank.csv",
			1, 1, 0, previewBytes, []byte(`{}`), nil, time.Now(), time.Now(),
		))
	mock.ExpectQuery(`SELECT (.+) FROM "quiz_questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"question_id", "question", "type", "options", "answers"}).
			AddRow(existingQuestionID.String(), "Capital?", 1, `{"1":"Paris","2":"Berlin"}`, `[1]`))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit", ctrl.CommitQuestionImportJob)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/import-jobs/"+jobID.String()+"/commit", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}
}
