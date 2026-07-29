package middlewares

import (
	"net/http"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	kratosClient "github.com/randhir3-cloud/GK-Circle-v2/api/pkg/kratos"
)

// VerificationRequired rejects authenticated requests from identities whose
// primary email address is not yet verified.
//
// Must be chained AFTER KratosAuthenticated, which already validates the
// Kratos session and loads the identity into c.Locals(constants.KratosUserDetails).
//
// Returns 401 AUTHENTICATION_REQUIRED if auth context is missing.
// Returns 403 EMAIL_VERIFICATION_REQUIRED if identity is unverified.
func (m *Middleware) VerificationRequired(c *fiber.Ctx) error {
	kratosUser, ok := c.Locals(constants.KratosUserDetails).(config.KratosUserDetails)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"code":    "AUTHENTICATION_REQUIRED",
			"message": "Authentication is required.",
		})
	}

	if !kratosClient.IsEmailVerified(kratosUser) {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"code":    "EMAIL_VERIFICATION_REQUIRED",
			"message": "Verify your email address to continue.",
		})
	}

	return c.Next()
}
