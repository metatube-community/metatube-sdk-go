package convertor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceSpaceAll(t *testing.T) {
	for _, unit := range []struct {
		origin, expect string
	}{
		{"Hello, world!", "Hello,world!"},
		{"Hello,\tworld!", "Hello,world!"},
		{"\t\tHe\tllo, \tworld!  \t", "Hello,world!"},
	} {
		assert.Equal(t, unit.expect, ReplaceSpaceAll(unit.origin))
	}
}

func TestNormalizeNewlines(t *testing.T) {
	for _, unit := range []struct {
		origin, expect string
	}{
		{"", ""},
		{"Hello, world!", "Hello, world!"},
		{"Hello,\r\nworld!", "Hello,\nworld!"},
		{"Hello,\rworld!", "Hello,\nworld!"},
		{"\r\n  Hello,\r\n\r\nworld!\r\n ", "Hello,\n\nworld!"},
	} {
		assert.Equal(t, unit.expect, NormalizeNewlines(unit.origin))
	}
}
