package javbus

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJavBus_GetMovieInfoByID(t *testing.T) {
	provider := NewJavBus()
	for _, item := range []string{
		"SSNI-776",
		"ABP-331",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.Nil(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestJavBus_SearchMovie(t *testing.T) {
	provider := NewJavBus()
	for _, item := range []string{
		"SSIS-033",
		"MIDE-154",
	} {
		results, err := provider.SearchMovie(item)
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.Nil(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
