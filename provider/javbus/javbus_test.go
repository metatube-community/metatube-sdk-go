package javbus

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJavBus_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"SMBD-77",
		"SSNI-776",
		"ABP-331",
		"CEMD-232",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestJavBus_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"SSIS-033",
		"MIDV-005",
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
