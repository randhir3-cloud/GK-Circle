package utils

import (
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderRequestID     = "X-Request-ID"
	maxCorrelationIDLen = 128
)

// ResolveAuditCorrelationID returns a correlation identifier for audit rows.
// Preference order: X-Correlation-ID, X-Request-ID, then a newly generated UUID.
func ResolveAuditCorrelationID(c *fiber.Ctx) string {
	candidates := []string{
		c.Get(HeaderCorrelationID),
		c.Get(HeaderRequestID),
	}
	for _, candidate := range candidates {
		normalized := normalizeCorrelationID(candidate)
		if normalized != "" {
			return normalized
		}
	}
	return uuid.NewString()
}

func normalizeCorrelationID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) > maxCorrelationIDLen {
		return trimmed[:maxCorrelationIDLen]
	}
	return trimmed
}
