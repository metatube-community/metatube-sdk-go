package collections

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedSet(t *testing.T) {
	set := NewOrderedSet(func(v int) string {
		return strconv.Itoa(v)
	})

	set.Add(1, 4, 6, 8, 9)
	set.Add(7, 4, 9, 2, 3)
	assert.Equal(t, 8, set.Len())
	assert.Equal(t, []int{1, 4, 6, 8, 9, 7, 2, 3}, set.Slice())

	set.Del(4, 5, 6, 7)
	assert.Equal(t, 5, set.Len())
	assert.Equal(t, []int{1, 8, 9, 2, 3}, set.Slice())

	b, _ := json.Marshal(set)
	assert.JSONEq(t, `[1,8,9,2,3]`, string(b))

	set2 := NewOrderedSet(func(v int) string {
		return strconv.Itoa(v)
	})
	_ = json.Unmarshal(b, set2)
	assert.Equal(t, 5, set.Len())
	assert.Equal(t, []int{1, 8, 9, 2, 3}, set.Slice())
}
