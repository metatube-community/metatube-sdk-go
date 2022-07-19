package jav321

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJAV321_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"heyzo2818",
		"300maan-791",
		"sivr00215",
		"ebod00916",
		"118abp00559",
		"nima00011",
		"pred00402",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestJAV321_SearchMovie(t *testing.T) {
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
