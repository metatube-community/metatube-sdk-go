package gfriends

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGFriends_GetActorInfoByID(t *testing.T) {
	provider := New().(*GFriends)
	for _, item := range []string{
		"小松凛花",
		"谷あづさ",
	} {
		info, err := provider.GetActorInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestGFriends_SearchActor(t *testing.T) {
	provider := New().(*GFriends)
	for _, item := range []string{
		"美竹すず",
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
