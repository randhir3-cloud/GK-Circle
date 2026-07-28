package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/middlewares"
)

func TestInstructorAnalytics_UnauthenticatedAccess_Returns401(t *testing.T) {
	app := fiber.New()
	m := middlewares.Middleware{}

	app.Get("/instructor/analytics/overview", m.EnsureCorrelationID, m.KratosAuthenticated, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/instructor/analytics/overview", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for unauthenticated request, got %d", resp.StatusCode)
	}
}

func TestInstructorAnalytics_NonOwnerAccess_Returns403(t *testing.T) {
	app := fiber.New()
	m := middlewares.Middleware{}

	// Simulate route protected by VerifyQuizAnalyticsAccess
	app.Get("/quizzes/:quiz_id/analytics/summary", func(c *fiber.Ctx) error {
		// Set permission to empty or non-owner
		c.Locals(constants.ContextQuizPermission, "")
		return m.VerifyQuizAnalyticsAccess(c)
	}, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/quizzes/11111111-1111-1111-1111-111111111111/analytics/summary", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-owner access, got %d", resp.StatusCode)
	}
}

func TestInstructorAnalytics_OwnerAccess_Allowed(t *testing.T) {
	app := fiber.New()
	m := middlewares.Middleware{}

	// Simulate route with SharePermission (owner)
	app.Get("/quizzes/:quiz_id/analytics/summary", func(c *fiber.Ctx) error {
		c.Locals(constants.ContextQuizPermission, constants.SharePermission)
		return m.VerifyQuizAnalyticsAccess(c)
	}, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/quizzes/11111111-1111-1111-1111-111111111111/analytics/summary", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for owner access, got %d", resp.StatusCode)
	}
}
