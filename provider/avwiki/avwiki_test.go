package avwiki

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAVWiki_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ABP-588",
		"PGD-919",
		"ORECO-062",
		"RECEN-012",
		"DDH-079",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestAVWiki_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ABP-588",
		"ORECO-062",
		"AKDL-030",
		"SABA-099",
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
