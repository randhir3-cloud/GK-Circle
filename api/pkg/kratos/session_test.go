package kratos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAuthenticatedUserRequiresBrowserCookieHeader(t *testing.T) {
	_, status, err := GetAuthenticatedUser("http://kratos.invalid", "")
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestGetAuthenticatedUserForwardsCookieToWhoAmI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/sessions/whoami" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if request.Header.Get("Cookie") != "ory_kratos_session=browser-session" {
			t.Fatalf("cookie header was not forwarded")
		}
		response.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(response).Encode(map[string]any{
			"identity": map[string]any{
				"id": "identity-1",
				"traits": map[string]any{
					"email": "learner@example.com",
					"name":  map[string]any{"first": "Learner", "last": "One"},
				},
			},
		})
	}))
	defer server.Close()

	user, status, err := GetAuthenticatedUser(server.URL, "ory_kratos_session=browser-session")
	if err != nil {
		t.Fatalf("GetAuthenticatedUser() error = %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if user.Identity.ID != "identity-1" {
		t.Fatalf("identity id = %q", user.Identity.ID)
	}
}
