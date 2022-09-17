package h4610

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestH4610_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"tk0047",
		"pla0051",
		"tk0062",
		"tk0050",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
