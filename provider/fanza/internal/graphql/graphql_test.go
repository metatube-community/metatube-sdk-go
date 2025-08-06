package graphql

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildQueryOptions(t *testing.T) {
	tests := []struct {
		inputURL string
		expected ContentPageDataQueryOptions
	}{
		{"https://video.dmm.co.jp/anime/content/?abc=123", ContentPageDataQueryOptions{IsAnime: true}},
		{"https://video.dmm.co.jp/amateur/content/?id=smjx065", ContentPageDataQueryOptions{IsAmateur: true}},
		{"https://video.dmm.co.jp/cinema/content/", ContentPageDataQueryOptions{IsCinema: true}},
		{"https://video.dmm.co.jp/av/content/?abc=123", ContentPageDataQueryOptions{IsAv: true}},
		{"https://video.dmm.co.jp/vr/content/?abc=123", ContentPageDataQueryOptions{IsAv: true}},
		{"https://video.dmm.co.jp/unknown/content/?abc=123", ContentPageDataQueryOptions{IsAv: true}}, // default case
	}

	for _, tt := range tests {
		opts := BuildContentPageDataQueryOptions(tt.inputURL)
		assert.Equal(t, tt.expected, opts)
	}
}

func TestClient_GetContentPageData(t *testing.T) {
	client := NewClient(WithHTTPClient(http.DefaultClient))
	client.gc.Log = func(s string) { t.Log(s) }

	content, err := client.GetContentPageData("1start00190", ContentPageDataQueryOptions{IsAv: true})
	require.NoError(t, err)
	require.NotNil(t, content)

	text, _ := json.MarshalIndent(content, "", "\t")
	t.Log(string(text))
}

func TestClient_GetContentPageData_Error(t *testing.T) {
	client := NewClient(WithHTTPClient(http.DefaultClient))
	client.gc.Log = func(s string) { t.Log(s) }

	_, err := client.GetContentPageData("oj8k666", ContentPageDataQueryOptions{IsAv: true})
	require.ErrorIs(t, err, ErrNullResponse)
}

func TestClient_GetUserReviews(t *testing.T) {
	client := NewClient(WithHTTPClient(http.DefaultClient))
	client.gc.Log = func(s string) { t.Log(s) }

	content, err := client.GetUserReviews("1start00190", 0)
	require.NoError(t, err)
	require.NotNil(t, content)

	text, _ := json.MarshalIndent(content, "", "\t")
	t.Log(string(text))
}

func TestClient_GetUserReviews_Error(t *testing.T) {
	client := NewClient(WithHTTPClient(http.DefaultClient))
	client.gc.Log = func(s string) { t.Log(s) }

	_, err := client.GetUserReviews("oj8k666", 0)
	require.ErrorIs(t, err, ErrNullResponse)
}
