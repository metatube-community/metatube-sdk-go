package dahlia

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDAHLIA_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"dldss265",
		"DLDSS264",
		"dldss087",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestDAHLIA_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"dldss-287",
		"DLDSS-259",
		"dldss271",
		"DLDSS274",
		"dldss087",
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
