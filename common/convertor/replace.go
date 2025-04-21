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
