package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	DATA "github.com/DATA-DOG/go-sqlmock"
	goqu "github.com/doug-martin/goqu/v9"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"go.uber.org/zap"
)

func TestQuizResultAdminController_GetReleaseStatus(t *testing.T) {
	sqlDB, mock, err := DATA.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	goquDB := goqu.New("postgres", sqlDB)
	adminService := services.NewQuizResultAdminService(goquDB, zap.NewNop())
	ctrl := NewQuizResultAdminController(adminService, zap.NewNop())

	quizID := uuid.New()
	now := time.Now().UTC()

	// EnsureScheduledReleaseEffective (future schedule → no-op)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Quiz", "Description", "owner-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "SCHEDULED", false, now.Add(2*time.Hour), nil, true, true, true, true))
	mock.ExpectCommit()

	// GetSelfPacedMetaByID
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Quiz", "Description", "owner-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "SCHEDULED", false, now.Add(2*time.Hour), nil, true, true, true, true))

	// Count assessment_attempts
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{"count"}).AddRow(int64(5)))

	app := fiber.New()
	app.Get("/quizzes/:quiz_id/results/status", ctrl.GetReleaseStatus)

	req := httptest.NewRequest(http.MethodGet, "/quizzes/"+quizID.String()+"/results/status", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d expected 200 body=%s", resp.StatusCode, raw)
	}

	raw, _ := io.ReadAll(resp.Body)
	var jsonResp map[string]interface{}
	_ = json.Unmarshal(raw, &jsonResp)
	data, _ := jsonResp["data"].(map[string]interface{})

	if data["total_submitted_attempts"] != 5.0 {
		t.Fatalf("expected total_submitted_attempts = 5, got %v", data["total_submitted_attempts"])
	}
	if data["is_currently_released"] != false {
		t.Fatalf("expected is_currently_released = false for future scheduled quiz, got %v", data["is_currently_released"])
	}
}

func TestQuizResultAdminController_UpdateResultSettings_RequiresScheduledDate(t *testing.T) {
	sqlDB, mock, err := DATA.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	goquDB := goqu.New("postgres", sqlDB)
	adminService := services.NewQuizResultAdminService(goquDB, zap.NewNop())
	ctrl := NewQuizResultAdminController(adminService, zap.NewNop())

	quizID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Quiz", "Description", "owner-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "IMMEDIATE", true, nil, now, true, true, true, true))
	mock.ExpectRollback()

	app := fiber.New()
	app.Patch("/quizzes/:quiz_id/results/settings", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "owner-1", Email: "owner@example.com"})
		return c.Next()
	}, ctrl.UpdateResultSettings)

	reqBody, _ := json.Marshal(structs.UpdateQuizResultSettingsRequest{
		ResultReleasePolicy: structs.ResultReleasePolicyScheduled,
		ResultsScheduledAt:  nil, // Missing date
	})

	req := httptest.NewRequest(http.MethodPatch, "/quizzes/"+quizID.String()+"/results/settings", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 400 when SCHEDULED missing date, got %d body=%s", resp.StatusCode, raw)
	}
}

func TestQuizResultAdminController_ReleaseResults_RejectsImmediatePolicy(t *testing.T) {
	sqlDB, mock, err := DATA.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	goquDB := goqu.New("postgres", sqlDB)
	adminService := services.NewQuizResultAdminService(goquDB, zap.NewNop())
	ctrl := NewQuizResultAdminController(adminService, zap.NewNop())

	quizID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(quizID, "BPSC Quiz", "Description", "owner-1", true, "SELF_PACED", "PUBLISHED", int64(1800), 1, 0.0, true, "IMMEDIATE", true, nil, now, true, true, true, true))
	mock.ExpectRollback()

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/results/release", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "owner-1", Email: "owner@example.com"})
		return c.Next()
	}, ctrl.ReleaseResults)

	req := httptest.NewRequest(http.MethodPost, "/quizzes/"+quizID.String()+"/results/release", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 400 when releasing IMMEDIATE policy, got %d body=%s", resp.StatusCode, raw)
	}
}
