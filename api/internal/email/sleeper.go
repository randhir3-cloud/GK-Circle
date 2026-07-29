package email

import (
	"context"
	"time"
)

// Sleeper defines an interface for mockable execution delays respecting request contexts.
type Sleeper interface {
	Sleep(ctx context.Context, delay time.Duration) error
}

// ContextSleeper implements Sleeper using context selection and time.Timer.
type ContextSleeper struct{}

// Sleep blocks execution for the specified duration or returns early if the context is cancelled.
func (ContextSleeper) Sleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
