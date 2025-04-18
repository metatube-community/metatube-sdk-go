package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	k := [][]int{{1, 2}, {3, 4}}
	assert.Equal(t, []int{1, 2, 3, 4}, Flatten(k))

	s := [][][]string{{{"a", "b"}, {"c", "d"}}, {{"e", "f"}, {"g", "h"}}}
	assert.Equal(t, [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}, {"g", "h"}}, Flatten(s))
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, Flatten(Flatten(s)))
}
