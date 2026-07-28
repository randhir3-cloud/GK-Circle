package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
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
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

func TestCreateTestSnapshot_RejectsUnresolvedDynamicCollection(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitTestSnapshotController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	collectionID := uuid.New()
	now := time.Now().UTC()
	filterJSON := []byte(`{"difficulty":"hard"}`)

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "Hard", "DYNAMIC", 0, filterJSON, "editor-1", now, now))
	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "Hard", "DYNAMIC", 0, filterJSON, "editor-1", now, now))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/test-snapshots", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUid, "editor-1")
		return c.Next()
	}, ctrl.CreateSnapshot)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/test-snapshots", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}
}

func TestGetLearnerTestSnapshot_OmitsAnswerKeys(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitTestSnapshotController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	quizID := uuid.New()
	snapshotID := uuid.New()
	questionID := uuid.New()
	lineageID := uuid.New()
	now := time.Now().UTC()
	options, _ := json.Marshal(map[string]string{"1": "A"})
	answers, _ := json.Marshal([]int{1})
	sourceJSON, _ := json.Marshal([]uuid.UUID{uuid.New()})

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
			int16(5), 30, "text", "text", nil, now,
		))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/test-snapshots/:snapshot_id/learner", ctrl.GetLearnerSnapshot)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/test-snapshots/"+snapshotID.String()+"/learner", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, body)
	}
	body, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Data structs.TestSnapshotLearnerResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, body)
	}
	if len(envelope.Data.Items) != 1 {
		t.Fatalf("items = %d", len(envelope.Data.Items))
	}
	if regexp.MustCompile(`"answers"|"official_answer"|"authoritative_answer"|"answer_review_status"`).Match(body) {
		t.Fatalf("learner payload leaked keys: %s", body)
	}
}
