package util

import (
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// ParseDate parses a string with valid date format into time.Time.
func ParseDate(s string) time.Time {
	t, _ := dateparse.ParseAny(s)
	return t
}

// ParseDuration parses a string with valid duration format into time.Duration.
func ParseDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "分", "m", 1)
	s = strings.Replace(s, "min", "m", 1)
	d, _ := time.ParseDuration(s)
	return d
}
