package m3u8

import (
	"testing"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

func TestParseMediaURI(t *testing.T) {
	url := "https://ppvclips02.aventertainments.com/01m3u8/mmdv-120/mmdv-120.m3u8"
	resp, _ := fetch.Fetch(url)
	defer resp.Body.Close()
	t.Log(ParseMediaURI(resp.Body))
}
