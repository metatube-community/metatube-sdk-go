package avleague

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAVLeague_GetActorInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"8301",
		"14005",
		"36672",
		"32759",
		"36736",
	} {
		info, err := provider.GetActorInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestAVLeague_SearchActor(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"白川ゆず",
		"美竹すず",
		"宇流木さら",
	} {
		results, err := provider.SearchActor(item)
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
