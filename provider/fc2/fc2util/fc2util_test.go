package fc2util

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNumber(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want string
	}{
		{"738573", "738573"},
		{"FC2-738573", "738573"},
		{"FC2-738573", "738573"},
		{"FC2 738573", "738573"},
		{"FC2_738573", "738573"},
		{"FC2-PPV-738573", "738573"},
		{"FC2 PPV 738573", "738573"},
		{"FC2PPV-738573", "738573"},
		{"FC2_PPV738573", "738573"},
		{"FC2PPV738573", "738573"},
		{"FC2PPV_738573", "738573"},
		// invalid cases:
		{"Unknow", ""},
		{"Unknow 12345", ""},
		{"FC2 WRONG 12345", ""},
		{"FC3-PPV-12345", ""},
	} {
		assert.Equal(t, unit.want, ParseNumber(unit.orig), unit.orig)
	}
}

func TestFetchImage(t *testing.T) {
	for _, unit := range []struct {
		url  string
		want []byte
	}{
		{"https://storage89000.contents.fc2.com/file/392/39188234/1702896500.66.jpg", noImageContent},
	} {
		resp, err := FetchImage(unit.url)
		if assert.NoError(t, err) {
			data, _ := io.ReadAll(resp.Body)
			assert.Equal(t, unit.want, data)
		}
	}
}
