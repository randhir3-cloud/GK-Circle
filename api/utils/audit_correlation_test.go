package utils

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
)

func TestResolveAuditCorrelationID_PrefersCorrelationHeader(t *testing.T) {
	app := fiber.New()
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		resolved = ResolveAuditCorrelationID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(HeaderCorrelationID, " corr-123 ")
	req.Header.Set(HeaderRequestID, "req-999")
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resolved != "corr-123" {
		t.Fatalf("expected corr-123, got %q", resolved)
	}
}

func TestResolveAuditCorrelationID_FallsBackToRequestID(t *testing.T) {
	app := fiber.New()
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		resolved = ResolveAuditCorrelationID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, "req-abc")
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resolved != "req-abc" {
		t.Fatalf("expected req-abc, got %q", resolved)
	}
}

func TestResolveAuditCorrelationID_GeneratesWhenMissing(t *testing.T) {
	app := fiber.New()
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		resolved = ResolveAuditCorrelationID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resolved == "" {
		t.Fatal("expected generated correlation id")
	}
}

func TestResolveAuditCorrelationID_TruncatesLongValues(t *testing.T) {
	app := fiber.New()
	var resolved string
	app.Get("/", func(c *fiber.Ctx) error {
		resolved = ResolveAuditCorrelationID(c)
		return c.SendStatus(fiber.StatusOK)
	})

	long := stringsRepeat("a", maxCorrelationIDLen+40)
	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(HeaderCorrelationID, long)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if len(resolved) != maxCorrelationIDLen {
		t.Fatalf("expected length %d, got %d", maxCorrelationIDLen, len(resolved))
	}
}

func stringsRepeat(value string, count int) string {
	out := make([]byte, 0, count*len(value))
	for i := 0; i < count; i++ {
		out = append(out, value...)
	}
	return string(out)
}
