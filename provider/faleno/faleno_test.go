package faleno

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFALENO_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"fsdss754",
		"FSDSS749",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestFALENO_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"FSDSS-723",
		"FSDSS746",
		"fsdss728",
		"fsdss-721",
	} {
		results, err := provider.SearchMovie(provider.NormalizeMovieKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
