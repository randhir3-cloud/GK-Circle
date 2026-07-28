package v1

import (
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
	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"go.uber.org/zap"
)

func newLearnerAnalyticsTestApp(t *testing.T) (*fiber.App, DATA.Sqlmock, *LearnerAnalyticsController) {
	t.Helper()
	sqlDB, mock, err := DATA.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db := goqu.New("postgres", sqlDB)
	svc := services.NewLearnerAnalyticsAggregationService(db, services.NewLearnerAnalyticsCache(nil, zap.NewNop()), zap.NewNop())
	ctrl := NewLearnerAnalyticsControllerWithService(svc, zap.NewNop())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUser, models.User{ID: "learner-1", Email: "l@example.com"})
		return c.Next()
	})
	app.Get("/analytics/dashboard", ctrl.GetDashboard)
	app.Get("/analytics/activity", ctrl.GetRecentActivity)
	app.Get("/analytics/trends", ctrl.GetPerformanceTrends)
	app.Get("/analytics/subjects", ctrl.GetSubjectPerformance)
	app.Get("/analytics/attempts/:attempt_id/timeline", ctrl.GetAttemptTimeline)
	return app, mock, ctrl
}

func TestGetDashboardSummary_Success(t *testing.T) {
	app, mock, _ := newLearnerAnalyticsTestApp(t)
	quizID := uuid.New()
	attemptID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "quiz_id", "user_id", "status", "total_score", "max_score", "time_taken_seconds",
			"submitted_at", "created_at", "quiz_title", "category_id", "category_name", "results_visible",
		}).AddRow(attemptID, quizID, "learner-1", "SUBMITTED", 8.0, 10.0, int64(120), now, now, "Quiz", nil, nil, true))

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_analytics_events"`).
		WillReturnRows(DATA.NewRows([]string{"attempt_id", "metadata"}))

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/analytics/dashboard", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d body=%s", resp.StatusCode, raw)
	}
	raw, _ := io.ReadAll(resp.Body)
	var envelope map[string]interface{}
	_ = json.Unmarshal(raw, &envelope)
	data := envelope["data"].(map[string]interface{})
	if data["total_attempts"].(float64) != 1 {
		t.Fatalf("total_attempts=%v", data["total_attempts"])
	}
	if data["resolved_timezone"] == nil || data["resolved_timezone"] == "" {
		t.Fatal("missing resolved_timezone")
	}
}

func TestGetRecentActivity_OpaqueCursorPagination(t *testing.T) {
	app, mock, _ := newLearnerAnalyticsTestApp(t)
	quizID := uuid.New()
	now := time.Now().UTC()
	a1 := uuid.New()
	a2 := uuid.New()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "quiz_id", "user_id", "status", "total_score", "max_score", "time_taken_seconds",
			"submitted_at", "created_at", "quiz_title", "category_id", "category_name", "results_visible",
		}).AddRow(a1, quizID, "learner-1", "SUBMITTED", 5.0, 10.0, int64(60), now, now, "Quiz A", nil, nil, true).
			AddRow(a2, quizID, "learner-1", "IN_PROGRESS", nil, nil, nil, nil, now.Add(-time.Hour), "Quiz B", nil, nil, false))

	req := httptest.NewRequest(http.MethodGet, "/analytics/activity?limit=1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d body=%s", resp.StatusCode, raw)
	}
	raw, _ := io.ReadAll(resp.Body)
	var envelope map[string]interface{}
	_ = json.Unmarshal(raw, &envelope)
	data := envelope["data"].(map[string]interface{})
	items := data["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("items=%d", len(items))
	}
	if data["has_more"] != true {
		t.Fatal("expected has_more")
	}
	if data["next_cursor"] == nil || data["next_cursor"] == "" {
		t.Fatal("expected next_cursor")
	}
}

func TestGetPerformanceTrends_DateRangeLimits(t *testing.T) {
	app, mock, _ := newLearnerAnalyticsTestApp(t)
	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "quiz_id", "user_id", "status", "total_score", "max_score", "time_taken_seconds",
			"submitted_at", "created_at", "quiz_title", "category_id", "category_name", "results_visible",
		}))

	from := time.Now().UTC().AddDate(0, 0, -2).Format(time.RFC3339)
	to := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/analytics/trends?granularity=daily&from="+from+"&to="+to, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d body=%s", resp.StatusCode, raw)
	}
	raw, _ := io.ReadAll(resp.Body)
	var envelope map[string]interface{}
	_ = json.Unmarshal(raw, &envelope)
	data := envelope["data"].(map[string]interface{})
	buckets := data["buckets"].([]interface{})
	if len(buckets) < 2 {
		t.Fatalf("expected complete buckets, got %d", len(buckets))
	}
	first := buckets[0].(map[string]interface{})
	if _, ok := first["average_percentage"]; !ok {
		t.Fatal("average_percentage missing")
	}
	if first["average_percentage"] != nil {
		t.Fatal("empty bucket average_percentage must be null")
	}
}

func TestLearnerAnalytics_ForbiddenAccessOnOtherUserAttempt(t *testing.T) {
	app, mock, _ := newLearnerAnalyticsTestApp(t)
	attemptID := uuid.New()
	quizID := uuid.New()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT (.+) FROM "assessment_attempts"`).
		WillReturnRows(DATA.NewRows([]string{
			"id", "quiz_id", "user_id", "status", "total_score", "max_score", "time_taken_seconds",
			"submitted_at", "created_at", "quiz_title", "category_id", "category_name", "results_visible",
		}).AddRow(attemptID, quizID, "other-user", "SUBMITTED", 9.0, 10.0, int64(90), now, now, "Quiz", nil, nil, true))

	req := httptest.NewRequest(http.MethodGet, "/analytics/attempts/"+attemptID.String()+"/timeline", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403, got %d body=%s", resp.StatusCode, raw)
	}
}
