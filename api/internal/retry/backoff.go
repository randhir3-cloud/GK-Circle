package retry

import (
	"math"
	"time"
)

// CalculateBackoff calculates the exponential backoff delay based on attempt number
func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}

	// base * 2^attempt
	multiplier := math.Pow(2, float64(attempt))
	delay := float64(baseDelay) * multiplier

	if delay > float64(maxDelay) {
		return maxDelay
	}

	return time.Duration(delay)
}
