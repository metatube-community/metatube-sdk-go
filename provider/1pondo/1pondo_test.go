package onepondo

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestOnePondo_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"071319_870",
		"042922_001",
		"080812_401",
		"071912_387",
		"050522_001",
	})
}

func TestOnePondo_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"071319_870",
		"071912_387",
	})
}
