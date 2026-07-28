package v1

import (
	"encoding/json"
	"io"
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
	"github.com/randhir3-cloud/GK-Circle-v2/api/middlewares"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/security"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

const leakSampleBody = `{"data":[{"correct_answer":"[1]","official_answer":"[1]","authoritative_answer":"[1]","answers":"[1]"}]}`

func newReviewProtectionApp(t *testing.T, userID string, editor *models.User) (*fiber.App, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	middleware := middlewares.NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))
	app := fiber.New()

	injectAuth := func(c *fiber.Ctx) error {
		c.Locals(constants.ContextUid, userID)
		if editor != nil {
			c.Locals(constants.ContextUser, *editor)
		}
		return c.Next()
	}

	leakHandler := func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Status(http.StatusOK).SendString(leakSampleBody)
	}

	app.Get("/analytics_board/user", injectAuth, middleware.VerifyPlayedQuizReviewAccess, leakHandler)
	app.Get("/final_score/user", injectAuth, middleware.VerifyPlayedQuizReviewAccess, leakHandler)
	app.Get("/user_played_quizes/:user_played_quiz_id", injectAuth, middleware.VerifyPlayedQuizReviewAccess, leakHandler)

	return app, mock
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return body
}

func expectReviewAccessQuery(mock sqlmock.Sqlmock, playedQuizID, quizID string) {
	mock.ExpectQuery(`SELECT (.+) FROM "user_played_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_admin_id", "quiz_id"}).
			AddRow("participant-1", "host-1", quizID))
}

func TestAnswerKeyProtection_UnauthenticatedReviewRoutesReject(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	middleware := middlewares.NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))
	app := fiber.New()
	app.Get("/analytics_board/user", middleware.VerifyPlayedQuizReviewAccess, func(c *fiber.Ctx) error {
		return c.SendString(leakSampleBody)
	})

	playedQuizID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/analytics_board/user?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if err := security.AssertNoSensitiveAnswerKeyFields(readBody(t, resp)); err != nil {
		t.Fatalf("unauthenticated response leaked answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_UnauthorizedCallerGetsNoAnswerKeys(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	app, mock := newReviewProtectionApp(t, "stranger-1", nil)
	expectReviewAccessQuery(mock, playedQuizID, quizID)

	req := httptest.NewRequest(http.MethodGet, "/analytics_board/user?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	if err := security.AssertNoSensitiveAnswerKeyFields(readBody(t, resp)); err != nil {
		t.Fatalf("forbidden response leaked answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_ParticipantReviewAllowsAnswerKeys(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	app, mock := newReviewProtectionApp(t, "participant-1", nil)
	expectReviewAccessQuery(mock, playedQuizID, quizID)

	req := httptest.NewRequest(http.MethodGet, "/user_played_quizes/"+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if err := security.AssertHasSensitiveAnswerKeyField(readBody(t, resp), "correct_answer"); err != nil {
		t.Fatalf("authorised participant review missing answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_HostReviewAllowsAnswerKeys(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	app, mock := newReviewProtectionApp(t, "host-1", nil)
	expectReviewAccessQuery(mock, playedQuizID, quizID)

	req := httptest.NewRequest(http.MethodGet, "/final_score/user?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if err := security.AssertHasSensitiveAnswerKeyField(readBody(t, resp), "correct_answer"); err != nil {
		t.Fatalf("authorised host review missing answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_EditorReviewAllowsAnswerKeys(t *testing.T) {
	playedQuizID := uuid.NewString()
	quizID := uuid.NewString()
	editor := &models.User{ID: "editor-1", Email: "editor@example.com"}
	app, mock := newReviewProtectionApp(t, "editor-1", editor)

	mock.ExpectQuery(`SELECT (.+) FROM "user_played_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "session_admin_id", "quiz_id"}).
			AddRow("participant-1", "host-1", quizID))
	mock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}))
	mock.ExpectQuery(`SELECT (.+) FROM "shared_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"permission"}).AddRow("read"))

	req := httptest.NewRequest(http.MethodGet, "/analytics_board/user?user_played_quiz="+playedQuizID, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if err := security.AssertHasSensitiveAnswerKeyField(readBody(t, resp), "correct_answer"); err != nil {
		t.Fatalf("authorised editor review missing answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_PublicQuizCatalogHasNoAnswerKeys(t *testing.T) {
	body := []byte(`{"data":[{"id":"quiz-1","title":"Demo","total_questions":5}]}`)
	if err := security.AssertNoSensitiveAnswerKeyFields(body); err != nil {
		t.Fatalf("public quiz catalog fixture leaked answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_LiveDeliveryPayloadHasNoAnswerKeys(t *testing.T) {
	question := models.Question{
		ID:                uuid.New(),
		QuizId:            uuid.New(),
		OrderNumber:       1,
		DurationInSeconds: 20,
		Question:          "Stem",
		Options:           map[string]string{"1": "A"},
		Answers:           []int{1},
		OfficialAnswer:    []int{1},
		AuthoritativeAnswer: []int{1},
	}
	payload := utils.BuildLiveQuestionDeliveryPayload(question, question.CreatedAt, 3, 4)
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := security.AssertNoSensitiveAnswerKeyFields(body); err != nil {
		t.Fatalf("live delivery payload leaked answer keys: %v", err)
	}
}

func TestAnswerKeyProtection_NestedLeakDetection(t *testing.T) {
	body := []byte(`{"meta":{"rows":[{"detail":{"authoritative_answer":"[2]"}}]}}`)
	if err := security.AssertNoSensitiveAnswerKeyFields(body); err == nil {
		t.Fatal("expected nested authoritative_answer leak to be detected")
	}
}
