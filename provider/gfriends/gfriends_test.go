package gfriends

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGFriends_GetActorInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"小澤マリア",
		"小松凛花",
		"谷あづさ",
		"若宮はずき",
	} {
		info, err := provider.GetActorInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestGFriends_GetActorInfoByURL(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"https://github.com/gfriends/gfriends?gfriends-id=%E5%B0%8F%E6%9D%BE%E5%87%9B%E8%8A%B1",
		"https://github.com/gfriends/gfriends?gfriends-id=%E8%B0%B7%E3%81%82%E3%81%A5%E3%81%95",
	} {
		info, err := provider.GetActorInfoByURL(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestGFriends_SearchActor(t *testing.T) {
	provider := New()
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
