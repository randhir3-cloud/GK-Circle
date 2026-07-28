package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
	"github.com/randhir3-cloud/GK-Circle-v2/api/security"
)

func TestImportPipelineRegression_RouteInventoryMatchesPhase3Surfaces(t *testing.T) {
	routes := security.CSVImportRoutes()
	names := map[string]bool{}
	for _, route := range routes {
		names[route.Name] = true
	}

	for _, required := range []string{
		"import_job_create_preview",
		"import_job_get_preview",
		"import_job_commit",
		"legacy_csv_upload",
	} {
		if !names[required] {
			t.Fatalf("missing route %s in inventory", required)
		}
	}
}

func TestImportPipelineRegression_GetPreviewJobAuthorized(t *testing.T) {
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
	preview := models.ImportPreviewPayload{
		ValidRows: []models.ImportPreviewRow{{
			RowNumber:           2,
			Question:            "Capital?",
			Type:                1,
			Options:             map[string]string{"1": "Paris", "2": "Berlin"},
			Answers:             []int{1},
			OfficialAnswer:      []int{1},
			AuthoritativeAnswer: []int{1},
			AnswerReviewStatus:  constants.AnswerReviewUnreviewed,
			RevisionNumber:      1,
		}},
	}
	previewBytes, _ := json.Marshal(preview)

	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}).AddRow(
			jobID, quizID, "editor-1", models.ImportJobStatusPreviewed, "bank.csv",
			1, 1, 0, previewBytes, []byte(`{}`), nil, time.Now(), time.Now(),
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/questions/import-jobs/:import_job_id", ctrl.GetQuestionImportJob)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/questions/import-jobs/"+jobID.String(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
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
	if len(envelope.Data.Preview.ValidRows) != 1 {
		t.Fatalf("valid rows = %d", len(envelope.Data.Preview.ValidRows))
	}
	if envelope.Data.Preview.ValidRows[0].AnswerReviewStatus != constants.AnswerReviewUnreviewed {
		t.Fatalf("authority status = %s", envelope.Data.Preview.ValidRows[0].AnswerReviewStatus)
	}
}

func TestImportPipelineRegression_GetPreviewJobWrongQuizNotFound(t *testing.T) {
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
	otherQuizID := uuid.New()
	jobID := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM "question_import_jobs"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "created_by", "status", "source_filename",
			"total_rows", "valid_row_count", "error_row_count", "preview_json",
			"commit_result_json", "committed_at", "created_at", "updated_at",
		}))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/questions/import-jobs/:import_job_id", ctrl.GetQuestionImportJob)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+otherQuizID.String()+"/questions/import-jobs/"+jobID.String(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	_ = quizID
}

func TestImportPipelineRegression_CommitAppliesAuthorityBeforeAppend(t *testing.T) {
	row := models.ImportPreviewRow{
		Question: "Stem",
		Type:     1,
		Options:  map[string]string{"1": "A", "2": "B"},
		Answers:  []int{2},
		Points:   5,
	}

	questions, err := models.QuestionsFromImportPreviewRows([]models.ImportPreviewRow{row})
	if err != nil {
		t.Fatalf("QuestionsFromImportPreviewRows: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("questions = %d", len(questions))
	}
	if questions[0].RevisionNumber != 1 {
		t.Fatalf("revision = %d", questions[0].RevisionNumber)
	}
	if questions[0].AnswerReviewStatus != constants.AnswerReviewUnreviewed {
		t.Fatalf("status = %s", questions[0].AnswerReviewStatus)
	}
	if len(questions[0].OfficialAnswer) != 1 || questions[0].OfficialAnswer[0] != 2 {
		t.Fatalf("official = %#v", questions[0].OfficialAnswer)
	}
}

func TestImportPipelineRegression_LegacyUploadRejectsDuplicates(t *testing.T) {
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

	mock.ExpectQuery(`SELECT (.+) FROM "quiz_questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"question_id", "question", "type", "options", "answers"}))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/questions/upload", func(c *fiber.Ctx) error {
		c.Locals(constants.FileName, csvPath)
		return c.Next()
	}, ctrl.ImportQuestionsByCsv)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/questions/upload", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}
}

func TestImportPipelineRegression_DuplicateClassificationKinds(t *testing.T) {
	rows := []models.ImportPreviewRow{
		{
			RowNumber: 2, Question: "Capital?", Type: 1,
			Options: map[string]string{"1": "Paris", "2": "Berlin"}, Answers: []int{1},
			OfficialAnswer: []int{1}, AuthoritativeAnswer: []int{1},
			AnswerReviewStatus: constants.AnswerReviewUnreviewed, RevisionNumber: 1,
		},
		{
			RowNumber: 4, Question: "Capital?", Type: 1,
			Options: map[string]string{"1": "Paris", "2": "Berlin"}, Answers: []int{1},
			OfficialAnswer: []int{1}, AuthoritativeAnswer: []int{1},
			AnswerReviewStatus: constants.AnswerReviewUnreviewed, RevisionNumber: 1,
		},
	}

	kept, errors := models.ApplyImportDuplicateDetection(rows, nil, models.ImportFingerprintIndex{})
	if len(kept) != 1 || len(errors) != 1 {
		t.Fatalf("kept=%d errors=%d", len(kept), len(errors))
	}
	if errors[0].Kind != models.ImportRowErrorKindDuplicate {
		t.Fatalf("kind = %s", errors[0].Kind)
	}
}
