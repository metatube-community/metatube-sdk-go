package m3u8

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
)

func TestParseBestMediaURI(t *testing.T) {
	url := "http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/sl.m3u8"
	resp, err := fetch.Fetch(url)
	if assert.NoError(t, err) {
		defer resp.Body.Close()
		t.Log(ParseBestMediaURI(resp.Body))
	}
}
