package airav

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAirAV_GetMovieInfoByID(t *testing.T) {
	provider := NewAirAV()
	for _, item := range []string{
		"STARS-381",
		//"DLDSS-077",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestAirAV_SearchMovie(t *testing.T) {
	provider := NewAirAV()
	for _, item := range []string{
		"STARS",
		"FC2",
		"DLDSS",
	} {
		results, err := provider.SearchMovie(item)
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
