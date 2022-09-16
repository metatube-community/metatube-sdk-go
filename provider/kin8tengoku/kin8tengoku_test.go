package kin8tengoku

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKIN8_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"3604",
		"3556",
		"3580",
		"3521",
		"3587",
		"1045",
		"3591",
		"3421",
		"3600",
		"2508",
		"1662",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
