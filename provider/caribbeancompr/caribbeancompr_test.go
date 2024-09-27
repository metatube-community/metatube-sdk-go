package caribbeancompr

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestCaribbeancomPremium_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"052121_002",
		"042922_001",
		"092018_010",
	})
}

func TestCaribbeancomPremium_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"062823_002",
	})
}
