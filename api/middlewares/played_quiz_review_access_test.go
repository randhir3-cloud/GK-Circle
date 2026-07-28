package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func newReviewAccessApp(t *testing.T, userID string) (*fiber.App, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	middleware := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))

	app := fiber.New()
	app.Get("/review", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUid, userID)
		return c.Next()
	}, middleware.VerifyPlayedQuizReviewAccess, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	return app, mock
}

func TestVerifyPlayedQuizReviewAccess_Unauthenticated(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	middleware := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))
	app := fiber.New()
	app.Get("/review", middleware.VerifyPlayedQuizReviewAccess, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	playedQuizID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/review?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestVerifyPlayedQuizReviewAccess_ForbiddenForStranger(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	app, mock := newReviewAccessApp(t, "stranger-1")

	mock.ExpectQuery(`SELECT (.+) FROM "user_played_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_admin_id", "quiz_id"}).
			AddRow("participant-1", "host-1", quizID))

	req := httptest.NewRequest(http.MethodGet, "/review?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestVerifyPlayedQuizReviewAccess_AllowsParticipant(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	app, mock := newReviewAccessApp(t, "participant-1")

	mock.ExpectQuery(`SELECT (.+) FROM "user_played_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_admin_id", "quiz_id"}).
			AddRow("participant-1", "host-1", quizID))

	req := httptest.NewRequest(http.MethodGet, "/review?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestVerifyPlayedQuizReviewAccess_AllowsHost(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	app, mock := newReviewAccessApp(t, "host-1")

	mock.ExpectQuery(`SELECT (.+) FROM "user_played_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_admin_id", "quiz_id"}).
			AddRow("participant-1", "host-1", quizID))

	req := httptest.NewRequest(http.MethodGet, "/review?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
