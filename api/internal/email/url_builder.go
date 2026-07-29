package email

import (
	"errors"
	"fmt"
	"net/url"
)

// AppURLBuilder generates trusted absolute application URLs.
type AppURLBuilder struct {
	baseURL *url.URL
}

// NewAppURLBuilder creates a validated AppURLBuilder instance.
func NewAppURLBuilder(baseURLStr string, appEnv string) (*AppURLBuilder, error) {
	if baseURLStr == "" {
		return nil, errors.New("WEB_URL configuration cannot be empty")
	}

	u, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid WEB_URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("WEB_URL scheme must be http or https")
	}

	if u.Host == "" {
		return nil, errors.New("WEB_URL host must be present")
	}

	if appEnv == "production" && u.Scheme != "https" {
		return nil, errors.New("production WEB_URL requires https scheme")
	}

	if u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("WEB_URL must not contain query parameters or fragments")
	}

	return &AppURLBuilder{baseURL: u}, nil
}

// isSafeIdentifier checks that the identifier contains only safe alphanumeric, dashes, or underscores.
func isSafeIdentifier(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// QuizInvitation constructs the URL for taking/joining a shared quiz.
func (b *AppURLBuilder) QuizInvitation(quizID, invitationID string) (string, error) {
	if !isSafeIdentifier(quizID) {
		return "", errors.New("invalid quiz identifier")
	}
	if !isSafeIdentifier(invitationID) {
		return "", errors.New("invalid invitation identifier")
	}

	u := *b.baseURL
	u.Path = fmt.Sprintf("/admin/quiz/list-quiz/%s", quizID)
	q := u.Query()
	q.Set("invitation", invitationID)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
