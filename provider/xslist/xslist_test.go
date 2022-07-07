package xslist

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXsList_GetActorInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"107",
		"24490",
		"5085",
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
		"白川ゆず",
		"果梨",
		"Saki",
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
