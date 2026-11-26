package convertor

import (
	"strings"
	"unicode"
)

// ReplaceSpaceAll removes all spaces in string.
func ReplaceSpaceAll(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, c := range s {
		if !unicode.IsSpace(c) {
			b.WriteRune(c)
		}
	}
	return b.String()
}

var newlineReplacer = strings.NewReplacer("\r\n", "\n", "\r", "\n")

// NormalizeNewlines converts CRLF and CR line endings to LF and
// trims leading and trailing spaces.
func NormalizeNewlines(s string) string {
	return strings.TrimSpace(newlineReplacer.Replace(s))
}
