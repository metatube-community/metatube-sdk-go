package prestige

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRESTIGE_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"hrv-014",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestPRESTIGE_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"edd-013",
	} {
		results, err := provider.SearchMovie(provider.TidyKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
