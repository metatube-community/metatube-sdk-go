package fc2hub

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFC2HUB_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"1152468-2725031",
		"1258474-3104947",
		"1258463-3104926",
		"1258427-3104805",
		"1258427-3104805",
		"230929-803681",
		"1259441-3106475",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestFC2HUB_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"FC2-PPV-2725031",
		"fc2-2417378",
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
