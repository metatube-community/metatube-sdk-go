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
