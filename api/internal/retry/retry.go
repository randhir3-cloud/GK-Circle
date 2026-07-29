package retry

import (
	"context"
	"time"
)

// Policy encapsulates the retry limits, delays, and backoff configuration.
type Policy struct {
	MaxAttempts    int
	RetryBaseDelay time.Duration
	RetryMaxDelay  time.Duration
}

// CalculateDelay computes the next retry backoff delay with jitter, or respects
// the provided Retry-After duration. In all cases, the delay is capped at RetryMaxDelay.
func (p Policy) CalculateDelay(attempt int, retryAfter time.Duration) time.Duration {
	// Attempt starts at 0 for the first request.
	// If attempt >= MaxAttempts-1, we shouldn't attempt any more retries.
	if attempt >= p.MaxAttempts-1 {
		return 0
	}

	var delay time.Duration
	if retryAfter > 0 {
		delay = retryAfter
	} else {
		delay = CalculateBackoff(attempt, p.RetryBaseDelay, p.RetryMaxDelay)
		delay = AddJitter(delay)
	}

	if delay > p.RetryMaxDelay {
		delay = p.RetryMaxDelay
	}

	return delay
}

// RetrySleeper provides an interface to execute mockable sleep actions inside retry loops.
type RetrySleeper interface {
	Sleep(ctx context.Context, delay time.Duration) error
}
