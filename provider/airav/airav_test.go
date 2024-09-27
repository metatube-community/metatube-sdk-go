package airav

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestAirAV_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"STARS-381",
		"FC2-PPV-2480488",
	})
}

func TestAirAV_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"ssni-278",
		"FC2-2735315",
	})
}
