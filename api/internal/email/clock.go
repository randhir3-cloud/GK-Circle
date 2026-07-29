package email

import "time"

// Clock defines an interface to retrieve the current time, facilitating mock testing.
type Clock interface {
	Now() time.Time
}

// SystemClock implements Clock using the standard time library in UTC.
type SystemClock struct{}

// Now returns the current time in UTC.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
