package v1

import (
	"bytes"
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
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

func TestCreateQuestionCollection_StaticReturns201(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitQuestionCollectionController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init controller: %v", err)
	}

	quizID := uuid.New()
	collectionID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectExec(`INSERT INTO "question_collections"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "Section A", "STATIC", 0, nil, "editor-1", now, now))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/collections", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUid, "editor-1")
		return c.Next()
	}, ctrl.CreateCollection)

	body, _ := json.Marshal(structs.ReqCreateQuestionCollection{
		Title: "Section A",
		Kind:  "STATIC",
	})
	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/collections", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, payload)
	}
}

func TestCreateQuestionCollection_StaticRejectsFilter(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitQuestionCollectionController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init controller: %v", err)
	}

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/collections", ctrl.CreateCollection)

	subject := "History"
	body, _ := json.Marshal(structs.ReqCreateQuestionCollection{
		Title:  "Section A",
		Kind:   "STATIC",
		Filter: &structs.CollectionDynamicFilter{Subject: &subject},
	})
	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+uuid.New().String()+"/collections", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestResolveQuestionCollection_ReturnsMetadataPending(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitQuestionCollectionController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init controller: %v", err)
	}

	quizID := uuid.New()
	collectionID := uuid.New()
	now := time.Now().UTC()
	filterJSON := []byte(`{"difficulty":"easy"}`)

	mock.ExpectQuery(`SELECT (.+) FROM "question_collections"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quiz_id", "title", "kind", "position", "filter_json", "created_by", "created_at", "updated_at",
		}).AddRow(collectionID, quizID, "Easy pool", "DYNAMIC", 0, filterJSON, "editor-1", now, now))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/collections/:collection_id/resolve", ctrl.ResolveCollection)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/collections/"+collectionID.String()+"/resolve", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body=%s", resp.StatusCode, payload)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var envelope struct {
		Data structs.CollectionResolutionResponse `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data.ResolutionStatus != models.CollectionResolutionStatusMetadataPending {
		t.Fatalf("status = %s", envelope.Data.ResolutionStatus)
	}
}
