package arzon

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestARZON_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"1291144",
		"1252925",
		"1624669",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestARZON_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"STARS",
		"IENF-209",
		"DLDSS-02",
		"FNEO-061",
	} {
		results, err := provider.SearchMovie(provider.NormalizeKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}

func TestARZON_Download(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"https://img.arzon.jp/image/1/1663/1663651L.jpg",
	} {
		resp, err := provider.Fetch(item)
		if assert.NoError(t, err) {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Log(b)
		}
	}
}
