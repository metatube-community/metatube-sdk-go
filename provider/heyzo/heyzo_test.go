package heyzo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeyzo_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"0841",
		"0805",
		"2189",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
