package email

import (
	"errors"
	"fmt"
	"time"
)

// ProviderErrorKind categorizes the root cause of an email provider failure.
type ProviderErrorKind string

const (
	ProviderErrorAuthentication ProviderErrorKind = "authentication"
	ProviderErrorRateLimited    ProviderErrorKind = "rate_limited"
	ProviderErrorTransient      ProviderErrorKind = "transient"
	ProviderErrorPermanent      ProviderErrorKind = "permanent"
	ProviderErrorTimeout        ProviderErrorKind = "timeout"
	ProviderErrorCancelled      ProviderErrorKind = "cancelled"
)

// ProviderError wraps provider errors with classified kinds, avoiding sensitive leakage.
type ProviderError struct {
	Provider   string
	Kind       ProviderErrorKind
	StatusCode int
	RetryAfter time.Duration
	Err        error
}

// Error returns a sanitized public message containing no credentials or recipient details.
func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf(
			"%s email provider failed with category %s and status %d",
			e.Provider,
			e.Kind,
			e.StatusCode,
		)
	}

	return fmt.Sprintf(
		"%s email provider failed with category %s",
		e.Provider,
		e.Kind,
	)
}

// Unwrap exposes the underlying error internally for diagnostic logging.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// Sentinel Validation Errors
var (
	ErrEmptyMessageID      = errors.New("message ID cannot be empty")
	ErrEmptyIdempotencyKey = errors.New("idempotency key cannot be empty")
	ErrEmptyRecipient      = errors.New("at least one recipient is required")
	ErrTooManyRecipients   = errors.New("recipient count exceeds maximum limit")
	ErrInvalidSender       = errors.New("invalid sender address")
	ErrInvalidReplyTo      = errors.New("invalid reply-to address")
	ErrEmptySubject        = errors.New("subject line cannot be empty")
	ErrSubjectTooLong      = errors.New("subject line exceeds 180 character limit")
	ErrCRLFInjection       = errors.New("carriage return or line feed detected in headers")
	ErrEmptyBody           = errors.New("both HTML and Plaintext bodies cannot be empty")
)
