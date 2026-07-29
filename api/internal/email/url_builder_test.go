package email

import (
	"testing"
)

func TestAppURLBuilder_NewValid(t *testing.T) {
	builder, err := NewAppURLBuilder("https://gkcircle.com", "production")
	if err != nil {
		t.Fatalf("Expected valid builder, got: %v", err)
	}

	if builder.baseURL.String() != "https://gkcircle.com" {
		t.Errorf("Expected base url 'https://gkcircle.com', got: %s", builder.baseURL.String())
	}
}

func TestAppURLBuilder_NewInvalidScheme(t *testing.T) {
	_, err := NewAppURLBuilder("ftp://gkcircle.com", "production")
	if err == nil {
		t.Error("Expected error for invalid scheme ftp, got nil")
	}
}

func TestAppURLBuilder_NewMissingHost(t *testing.T) {
	_, err := NewAppURLBuilder("https://", "production")
	if err == nil {
		t.Error("Expected error for missing host, got nil")
	}
}

func TestAppURLBuilder_NewProductionSSLRequired(t *testing.T) {
	_, err := NewAppURLBuilder("http://gkcircle.com", "production")
	if err == nil {
		t.Error("Expected error for http in production, got nil")
	}

	// Http is okay in development
	_, err = NewAppURLBuilder("http://localhost:3200", "development")
	if err != nil {
		t.Errorf("Expected valid HTTP builder in development, got error: %v", err)
	}
}

func TestAppURLBuilder_NewNoQueriesOrFragments(t *testing.T) {
	_, err := NewAppURLBuilder("https://gkcircle.com?foo=bar", "production")
	if err == nil {
		t.Error("Expected error for queries in base URL, got nil")
	}

	_, err = NewAppURLBuilder("https://gkcircle.com#fragment", "production")
	if err == nil {
		t.Error("Expected error for fragments in base URL, got nil")
	}
}

func TestAppURLBuilder_QuizInvitation(t *testing.T) {
	builder, err := NewAppURLBuilder("https://gkcircle.com", "production")
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	urlStr, err := builder.QuizInvitation("quiz-123", "inv-456")
	if err != nil {
		t.Fatalf("Failed to build quiz invitation: %v", err)
	}

	expected := "https://gkcircle.com/admin/quiz/list-quiz/quiz-123?invitation=inv-456"
	if urlStr != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, urlStr)
	}
}

func TestAppURLBuilder_QuizInvitationUnsafeIdentifiers(t *testing.T) {
	builder, err := NewAppURLBuilder("https://gkcircle.com", "production")
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	// Try traversal/injection
	_, err = builder.QuizInvitation("../malicious", "inv-123")
	if err == nil {
		t.Error("Expected error for traversal identifier, got nil")
	}

	_, err = builder.QuizInvitation("quiz-123", "inv-123&attacker=1")
	if err == nil {
		t.Error("Expected error for inject parameters, got nil")
	}
}
