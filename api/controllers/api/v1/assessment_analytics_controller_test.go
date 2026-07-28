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
	"go.uber.org/zap"
)

func TestAssessmentAnalyticsController_RejectsAuthoritativeClientEvents(t *testing.T) {
	sqlDB, mock, err := DATA.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	goquDB := goqu.New("postgres", sqlDB)
	ctrl := NewAssessmentAnalyticsController(goquDB, zap.NewNop())

	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	createdAt := time.Now().UTC().Add(-time.Hour)

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, "IN_PROGRESS",
			[]byte("[]"), 0.0, 10.0, createdAt, nil, nil, nil, nil, nil, createdAt, createdAt,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(
			quizID, "Quiz", "Desc", "owner-1", true, "SELF_PACED", "PUBLISHED",
			int64(1800), 1, 0.0, true, "IMMEDIATE", true, nil, nil, true, true, true, true,
		))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/attempts/:attempt_id/analytics/events", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1"})
		return c.Next()
	}, ctrl.RecordClientTelemetryBatch)

	body, _ := json.Marshal(structs.RecordTelemetryBatchRequest{
		Events: []structs.RecordTelemetryEventRequest{{
			EventType:  structs.EventAttemptStarted,
			OccurredAt: time.Now().UTC(),
			Metadata:   map[string]interface{}{},
		}},
	})
	req := httptest.NewRequest(
		http.MethodPost,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/analytics/events",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400 for authoritative client event, got %d body=%s", resp.StatusCode, raw)
	}
}

func TestAssessmentAnalyticsController_RejectsServerTelemetryClientEvents(t *testing.T) {
	sqlDB, mock, err := DATA.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	goquDB := goqu.New("postgres", sqlDB)
	ctrl := NewAssessmentAnalyticsController(goquDB, zap.NewNop())

	quizID := uuid.New()
	attemptID := uuid.New()
	snapshotID := uuid.New()
	createdAt := time.Now().UTC().Add(-time.Hour)

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "quiz_id", "user_id", "test_snapshot_id", "attempt_number", "status",
			"question_order", "negative_marks_per_question", "expected_max_score",
			"started_at", "submitted_at", "expires_at",
			"total_score", "max_score", "time_taken_seconds", "created_at", "updated_at",
		}).AddRow(
			attemptID, quizID, "learner-1", snapshotID, 1, "IN_PROGRESS",
			[]byte("[]"), 0.0, 10.0, createdAt, nil, nil, nil, nil, nil, createdAt, createdAt,
		))

	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "title", "description", "creator_id", "is_public", "assessment_mode", "status",
			"duration_seconds", "max_attempts", "negative_marks_per_question", "allow_answer_review",
			"result_release_policy", "results_released", "results_scheduled_at", "results_released_at",
			"show_score", "show_pass_fail", "show_correctness", "show_explanations",
		}).AddRow(
			quizID, "Quiz", "Desc", "owner-1", true, "SELF_PACED", "PUBLISHED",
			int64(1800), 1, 0.0, true, "IMMEDIATE", true, nil, nil, true, true, true, true,
		))

	app := fiber.New()
	app.Post("/quizzes/:quiz_id/attempts/:attempt_id/analytics/events", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1"})
		return c.Next()
	}, ctrl.RecordClientTelemetryBatch)

	body, _ := json.Marshal(structs.RecordTelemetryBatchRequest{
		Events: []structs.RecordTelemetryEventRequest{{
			EventType:  structs.EventResultViewed,
			OccurredAt: time.Now().UTC(),
			Metadata:   map[string]interface{}{},
		}},
	})
	req := httptest.NewRequest(
		http.MethodPost,
		"/quizzes/"+quizID.String()+"/attempts/"+attemptID.String()+"/analytics/events",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400 for server telemetry client event, got %d body=%s", resp.StatusCode, raw)
	}
}
