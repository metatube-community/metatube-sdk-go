package graphql

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPPVContent(t *testing.T) {
	client := NewClient()
	client.c.Log = func(s string) { t.Log(s) }

	content, err := client.GetPPVContent("1start00285v", QueryOption{IsAv: true})
	require.NoError(t, err)
	require.NotNil(t, content)

	text, _ := json.MarshalIndent(content, "", "\t")
	t.Log(string(text))
}
