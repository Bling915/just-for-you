package clock

import "time"

// Clock supplies time to application services without coupling them to the system clock.
type Clock interface {
	Now() time.Time
}

// System is the production clock.
type System struct{}

func (System) Now() time.Time { return time.Now() }
