package graphql

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateQueryOptions(t *testing.T) {
	tests := []struct {
		inputURL string
		expected QueryOptions
	}{
		{"https://video.dmm.co.jp/anime/content/?abc=123", QueryOptions{IsAnime: true}},
		{"https://video.dmm.co.jp/amateur/content/?id=smjx065", QueryOptions{IsAmateur: true}},
		{"https://video.dmm.co.jp/cinema/content/", QueryOptions{IsCinema: true}},
		{"https://video.dmm.co.jp/av/content/?abc=123", QueryOptions{IsAv: true}},
		{"https://video.dmm.co.jp/vr/content/?abc=123", QueryOptions{IsAv: true}},
		{"https://video.dmm.co.jp/unknown/content/?abc=123", QueryOptions{IsAv: true}}, // default case
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.inputURL)
		opts := GenerateQueryOptions(u)
		assert.Equal(t, tt.expected, opts)
	}
}

func TestGetPPVContent(t *testing.T) {
	client := NewClient()
	client.c.Log = func(s string) { t.Log(s) }

	content, err := client.GetPPVContent("1start00285v", QueryOptions{IsAv: true})
	require.NoError(t, err)
	require.NotNil(t, content)

	text, _ := json.MarshalIndent(content, "", "\t")
	t.Log(string(text))
}
