package heydouga

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeyDouga_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"heydouga-4037-479",
		"heydouga-4229-771",
		"heydouga-4229-759",
		"heydouga-4030-2000",
		"heydouga-4037-478",
	} {
		info, err := provider.GetMovieInfoByID(provider.NormalizeID(item))
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
