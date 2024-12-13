package ggjav

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestGGJAV_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"8130-1003647",
		"6604-1019534",
	})
}

func TestGGJAV_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"FC2-PPV-2725031",
		"fc2-2417378",
	})
}
