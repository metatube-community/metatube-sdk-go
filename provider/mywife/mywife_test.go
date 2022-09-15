package mywife

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMyWife_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"1252",
		"1341",
		"1542",
		"1882",
		"1888",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
