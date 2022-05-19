package xxx_av

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXxxAV_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"xxx-av-24719",
		"xxx-av-23395",
		"xxx-av-19337",
	} {
		info, err := provider.GetMovieInfoByID(provider.NormalizeID(item))
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
