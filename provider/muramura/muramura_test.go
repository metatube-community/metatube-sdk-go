package muramura

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestMuraMura_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"091522_959",
		"062509_011",
		"021110_163",
		"013010_157",
		"012810_155",
		"081222_953",
		"062509_003",
	})
}

func TestMuraMura_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		//"091522_959",
	})
}
