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
)

func newQuestionHTTPEnv(t *testing.T) (*fiber.App, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitQuestionController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init question controller: %v", err)
	}

	app := fiber.New()
	quizGroup := app.Group("/api/v1/quizzes/:" + constants.QuizId + "/questions")
	quizGroup.Get("/:"+constants.QuestionId+"/revisions", ctrl.ListQuestionRevisions)

	return app, mock
}

func TestListQuestionRevisions(t *testing.T) {
	app, mock := newQuestionHTTPEnv(t)

	quizID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5ea1")
	questionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	lineageID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eee")
	revisionID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eef")
	now := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT (.+) FROM "questions"`).
		WillReturnRows(sqlmock.NewRows([]string{"lineage_id", "revision_number"}).
			AddRow(lineageID, 2))

	mock.ExpectQuery(`SELECT (.+) FROM "question_revisions"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"question_id",
			"lineage_id",
			"revision_number",
			"question",
			"type",
			"options",
			"answers",
			"official_answer",
			"authoritative_answer",
			"answer_review_status",
			"answer_revision_reason",
			"answer_revision_source",
			"points",
			"duration_in_seconds",
			"question_media",
			"options_media",
			"resource",
			"created_by",
			"created_at",
		}).AddRow(
			revisionID,
			questionID,
			lineageID,
			2,
			"Stem?",
			1,
			`{"1":"A"}`,
			`[1]`,
			`[1]`,
			`[1]`,
			constants.AnswerReviewConfirmed,
			nil,
			nil,
			1,
			30,
			"text",
			"text",
			nil,
			nil,
			now,
		))

	path := "/api/v1/quizzes/" + quizID.String() + "/questions/" + questionID.String() + "/revisions"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data, ok := payload["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("data = %#v", payload["data"])
	}
	row, ok := data[0].(map[string]any)
	if !ok {
		t.Fatalf("row = %#v", data[0])
	}
	if int(row["revision_number"].(float64)) != 2 {
		t.Fatalf("revision_number = %#v", row["revision_number"])
	}
	if row["answer_review_status"] != constants.AnswerReviewConfirmed {
		t.Fatalf("answer_review_status = %#v", row["answer_review_status"])
	}
}
