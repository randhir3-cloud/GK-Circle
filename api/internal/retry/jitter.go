package retry

import (
	"math/rand"
	"time"
)

// AddJitter returns a random duration between [0, delay) using full jitter algorithm
func AddJitter(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(delay)))
}
