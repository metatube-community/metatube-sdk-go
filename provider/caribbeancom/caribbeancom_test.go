package caribbeancom

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaribbean_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"050422-001",
		"031222-001",
		"061014-618",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
