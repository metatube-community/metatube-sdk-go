package getchu

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetchu_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"4018339",
		"4042392",
		"4041955",
		"4042404",
		"4042423",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
