package middlewares

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	j "github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/jwt"
	kratosClient "github.com/randhir3-cloud/GK-Circle-v2/api/pkg/kratos"
	"github.com/randhir3-cloud/GK-Circle-v2/api/utils"
	"go.uber.org/zap"
)

func (m *Middleware) Authenticated(c *fiber.Ctx) error {
	token := c.Cookies(constants.CookieUser, "")
	kratosToken := c.Cookies(constants.KratosCookie, "")

	if kratosToken == "" {
		if token == "" {
			m.Logger.Debug("returning unauthorized after found empty token")
			return utils.JSONFail(c, http.StatusUnauthorized, constants.Unauthenticated)
		}
	} else {
		return m.KratosAuthenticated(c)
	}

	claims, err := jwt.ParseToken(m.Config, token)
	if err != nil {
		if errors.Is(err, j.ErrInvalidJWT()) || errors.Is(err, j.ErrTokenExpired()) {
			c.Cookie(RemoveCookie(constants.CookieUser))
			m.Logger.Error("JWT error during authentication in join", zap.Error(err))
			return utils.JSONFail(c, http.StatusUnauthorized, constants.Unauthenticated)
		}

		m.Logger.Error("error while checking user identity", zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUnauthenticated)
	}

	// Check if expiration time is less than 10 minutes than refesh token
	if time.Until(claims.Expiration()) < 10*time.Minute {
		cookieExpirationTime, err := time.ParseDuration(m.Config.Kratos.CookieExpirationTime)
		if err != nil {
			m.Logger.Error("unable to parse the duration for the cookie expiration", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrKratosCookieTime)
		}

		// Generate a new token
		newToken, err := jwt.CreateToken(m.Config, claims.Subject(), time.Now().Add(time.Hour*2))
		if err != nil {
			m.Logger.Error("error while refreshing token", zap.Error(err))
			return utils.JSONError(c, http.StatusInternalServerError, constants.ErrUnauthenticated)
		}

		// Set Coockie
		c.Cookie(&fiber.Cookie{
			Name:    constants.CookieUser,
			Value:   newToken,
			Expires: time.Now().Add(cookieExpirationTime),
		})
	}

	c.Locals(constants.ContextUid, claims.Subject())
	return c.Next()
}

func (m *Middleware) KratosAuthenticated(c *fiber.Ctx) error {
	kratosUser, status, err := kratosClient.GetAuthenticatedUser(
		m.Config.Kratos.BaseUrl,
		c.Get("Cookie"),
	)
	if err != nil {
		if m.Logger != nil {
			m.Logger.Debug("kratos session validation failed", zap.Error(err))
		}
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			return utils.JSONError(c, http.StatusUnauthorized, constants.ErrKratosIDEmpty)
		}
		return utils.JSONError(c, http.StatusBadGateway, constants.ErrKratosAuth)
	}

	m.Logger.Debug("userModel.GetUserByKratosID called", zap.Any("kratosID", kratosUser.Identity.ID))
	user, err := m.userModel.GetUserByKratosID(kratosUser.Identity.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			m.Logger.Error(constants.ErrGetUser, zap.Error(err))
			return utils.JSONError(c, http.StatusNotFound, constants.ErrGetUser)
		}
		m.Logger.Error(constants.ErrGetUser, zap.Error(err))
		return utils.JSONError(c, http.StatusInternalServerError, constants.ErrGetUser)
	}
	m.Logger.Debug("userModel.GetUserByKratosID success", zap.Any("user", user))

	c.Locals(constants.ContextUser, user)
	c.Locals(constants.ContextUid, user.ID)
	c.Locals(constants.KratosID, kratosUser.Identity.ID)
	c.Locals(constants.KratosUserDetails, kratosUser)
	return c.Next()

}
