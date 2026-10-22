package shared

import "time"

// Clock supplies time to application use cases without coupling them to the system clock.
type Clock interface {
	Now() time.Time
}
