package middlewares

import (
	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
)

const ContextCorrelationID = "correlation_id"

// EnsureCorrelationID resolves X-Correlation-ID / X-Request-ID (or generates one),
// stores it in locals, and echoes it on the response.
func (m Middleware) EnsureCorrelationID(c *fiber.Ctx) error {
	correlationID := utils.ResolveAuditCorrelationID(c)
	c.Locals(ContextCorrelationID, correlationID)
	c.Set(utils.HeaderCorrelationID, correlationID)
	return c.Next()
}
