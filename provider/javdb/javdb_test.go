package javdb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJavDB_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		//"BabyGotBoobs.18.12.27",
		//"BigNaturals.BigNaturals.23.05.27 Big Tits VR",
		//"BigNaturals Big Tits VR",
		//"BabyGotBoobs Lovely In Latex",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestJavBus_SearchMovie(t *testing.T) {
	//"http://127.0.0.1:8080/v1/movies/search?q=BigNaturals+Big+Tits+VR&provider=&fallback=True"
	provider := New()
	for _, item := range []string{
		"BigNaturals Big Tits VR",
		"BabyGotBoobs Lovely In Latex",
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
