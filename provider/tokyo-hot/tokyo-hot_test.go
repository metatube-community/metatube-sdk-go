package tokyohot

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokyoHot_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"s2mbd-002",
		"n1633",
		"n1624",
		"kb1624",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestTokyoHot_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"1624",
		"n0238",
	} {
		results, err := provider.SearchMovie(provider.NormalizeKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
