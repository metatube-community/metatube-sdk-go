package heydouga

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestHeyDouga_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"4037-479",
		"4030-1938",
		"4229-771",
		"4229-759",
		"4030-2000",
		"4037-478",
		"4226-032",
	})
}
