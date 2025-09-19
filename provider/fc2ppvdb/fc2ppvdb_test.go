package fc2ppvdb

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFC2PPVDB_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"2812904",
		"4669533",
		"4745474",
		"4137487",
	})
}
