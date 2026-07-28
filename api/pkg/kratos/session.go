package kratos

import (
	"errors"
	"net/http"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
)

var ErrUnauthenticated = errors.New("kratos session is not authenticated")

// GetAuthenticatedUser delegates cookie validation to Ory Kratos. The
// application never parses or reconstructs the Kratos session cookie.
func GetAuthenticatedUser(baseURL, cookieHeader string) (config.KratosUserDetails, int, error) {
	user := config.KratosUserDetails{}
	if strings.TrimSpace(cookieHeader) == "" {
		return user, http.StatusUnauthorized, ErrUnauthenticated
	}

	response, err := resty.New().
		SetBaseURL(strings.TrimRight(baseURL, "/")).
		R().
		SetHeader("Cookie", cookieHeader).
		SetHeader("Accept", "application/json").
		SetResult(&user).
		Get("/sessions/whoami")
	if err != nil {
		return user, http.StatusBadGateway, err
	}
	if response.StatusCode() != http.StatusOK {
		return user, response.StatusCode(), ErrUnauthenticated
	}

	return user, http.StatusOK, nil
}
