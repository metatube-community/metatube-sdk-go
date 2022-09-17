package c0930

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestC0930_NormalizeID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ki220913",
		"hitozuma1391",
		"hitozuma1371",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}
