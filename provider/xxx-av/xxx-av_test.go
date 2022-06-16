package xxx_av

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXxxAV_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"24719",
		"23395",
		"19337",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
