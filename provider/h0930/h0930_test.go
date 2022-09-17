package h0930

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestH0930_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ori1643",
		"ori1492",
		"ori1396",
		"orijuku823",
		"orimrs695",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
