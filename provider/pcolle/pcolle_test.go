package pcolle

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPcolle_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"14491760b933a35cfab",
		"156785614478ab480db",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
