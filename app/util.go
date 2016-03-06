package app

import (
	"time"
)

// Truncate a UTC time to the its date only
func date(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}
