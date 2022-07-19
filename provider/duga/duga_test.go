package duga

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDUGA_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"glory-4262",
		"waap-1294",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestDUGA_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"DINM",
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
