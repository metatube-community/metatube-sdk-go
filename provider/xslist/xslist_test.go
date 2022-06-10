package xslist

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXsList_GetActorInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"8",
		"12",
		"139",
		"175",
		"15659",
	} {
		info, err := provider.GetActorInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestXsList_SearchActor(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"Saki",
		"美竹すず",
		"川上ゆう（森野雫）",
		"新井エリー（晶エリー、大沢佑香）",
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
