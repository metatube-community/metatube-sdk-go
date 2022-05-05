package sod

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSOD_GetMovieInfoByID(t *testing.T) {
	provider := NewSOD()
	for _, item := range []string{
		"STARS-381",
		"DLDSS-077",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestSOD_SearchMovie(t *testing.T) {
	provider := NewSOD()
	for _, item := range []string{
		"STAR-399",
		"IENF-209",
		"DLDSS-02",
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
