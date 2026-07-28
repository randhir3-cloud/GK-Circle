package security

import (
	"testing"
)

func TestProtectedReviewRoutesInventory(t *testing.T) {
	routes := ProtectedReviewRoutes()
	if len(routes) != 3 {
		t.Fatalf("expected 3 protected review routes, got %d", len(routes))
	}

	for _, route := range routes {
		if route.Method == "" || route.Path == "" {
			t.Fatalf("invalid route inventory entry: %#v", route)
		}
		if !route.RequiresAuth || !route.RequiresReviewAccess {
			t.Fatalf("review route must require auth and review access: %#v", route)
		}
	}
}

func TestPreReleasePayloadRoutesInventory(t *testing.T) {
	routes := PreReleasePayloadRoutes()
	for _, route := range routes {
		if route.AllowsAnswerKeys {
			t.Fatalf("pre-release route must not allow answer keys: %#v", route)
		}
	}
}
