package fc2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFC2_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"2812904",
		"2676371",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
