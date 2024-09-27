package fc2hub

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFC2HUB_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"1152468-2725031",
		"1258474-3104947",
		"1258463-3104926",
		"1258427-3104805",
		"1258427-3104805",
		"230929-803681",
		"1259441-3106475",
	})
}

func TestFC2HUB_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"FC2-PPV-2725031",
		"fc2-2417378",
	})
}
