package fc2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFC2_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"406996",
		"2812904",
		"2676371",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestParseNumber(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want string
	}{
		{"738573", "738573"},
		{"FC2-738573", "738573"},
		{"FC2_738573", "738573"},
		{"FC2-PPV-738573", "738573"},
		{"FC2PPV-738573", "738573"},
		{"FC2_PPV738573", "738573"},
		{"FC2PPV738573", "738573"},
		{"FC2PPV_738573", "738573"},
	} {
		assert.Equal(t, unit.want, ParseNumber(unit.orig), unit.orig)
	}
}
