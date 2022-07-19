package mgstage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMGStage_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		//"TAP-002",
		"SIRO-2219",
		"300MAAN-778",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestMGStage_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ABP-177",
		"200GANA-2701",
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
