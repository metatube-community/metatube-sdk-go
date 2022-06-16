package heydouga

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeyDouga_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"4037-479",
		"4229-771",
		"4229-759",
		"4030-2000",
		"4037-478",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
