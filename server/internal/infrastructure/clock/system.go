package clock

import "time"

// System is the production clock.
type System struct{}

func (System) Now() time.Time { return time.Now() }
