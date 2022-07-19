package sod

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSOD_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"3DSVR-0416",
		//"DLDSS-077",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestSOD_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"STAR-399",
		"IENF-209",
		"DLDSS-02",
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

func TestSOD_Download(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"https://dy43ylo5q3vt8.cloudfront.net/_pics/202108/dldss_022/dldss_022_m.jpg",
	} {
		resp, err := provider.Fetch(item)
		if assert.NoError(t, err) {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Log(b)
		}
	}
}
