package avbase

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAVBase_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"prestige:ABP-588",
		"tameike:MEYD-856",
		"SSIS-354",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestAVBase_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ABP-588",
		"MEYD-856",
		"SSIS-354",
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
