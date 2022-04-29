package utility

import (
	"time"

	"github.com/araddon/dateparse"
)

// ParseDate parses a string with valid date format to time.Time.
func ParseDate(s string) time.Time {
	t, _ := dateparse.ParseAny(s)
	return t
}
