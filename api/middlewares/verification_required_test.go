package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"go.uber.org/zap"
)

// helper to build KratosUserDetails for tests
func makeTestKratosUser(email string, verified bool) config.KratosUserDetails {
	user := config.KratosUserDetails{}
	user.Identity.Traits.Email = email
	user.Identity.VerifiableAddresses = []struct {
		Value    string `json:"value"`
		Verified bool   `json:"verified"`
	}{
		{Value: email, Verified: verified},
	}
	return user
}

func TestVerificationRequired_Verified_CallsNext(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	t.Cleanup(func() { _ = sqlDB.Close() })
	m := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		user := makeTestKratosUser("test@gkcircle.com", true)
		c.Locals(constants.KratosUserDetails, user)
		return c.Next()
	}, m.VerificationRequired, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestVerificationRequired_Unverified_Returns403WithCode(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	t.Cleanup(func() { _ = sqlDB.Close() })
	m := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		user := makeTestKratosUser("test@gkcircle.com", false)
		c.Locals(constants.KratosUserDetails, user)
		return c.Next()
	}, m.VerificationRequired, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["code"] != "EMAIL_VERIFICATION_REQUIRED" {
		t.Fatalf("expected code EMAIL_VERIFICATION_REQUIRED, got %s", body["code"])
	}
}

func TestVerificationRequired_NoLocalsKey_Returns401(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	t.Cleanup(func() { _ = sqlDB.Close() })
	m := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))

	app := fiber.New()
	app.Get("/test", m.VerificationRequired, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["code"] != "AUTHENTICATION_REQUIRED" {
		t.Fatalf("expected code AUTHENTICATION_REQUIRED, got %s", body["code"])
	}
}

func TestVerificationRequired_WrongLocalsType_Returns401(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	t.Cleanup(func() { _ = sqlDB.Close() })
	m := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals(constants.KratosUserDetails, "wrong-type-string")
		return c.Next()
	}, m.VerificationRequired, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestVerificationRequired_MalformedVerificationData_Returns403(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	t.Cleanup(func() { _ = sqlDB.Close() })
	m := NewMiddleware(config.AppConfig{}, zap.NewNop(), goqu.New("postgres", sqlDB))

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		// User with no verifiable addresses or empty email
		c.Locals(constants.KratosUserDetails, config.KratosUserDetails{})
		return c.Next()
	}, m.VerificationRequired, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}
