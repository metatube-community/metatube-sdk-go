package aventertainments

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestAVE_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"4319",
		"7215",
		"142802",
		"9865",
		"10161",
		"12881",
		"140930",
		"115855",
		"142800",
	})
}

func TestAVE_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"lldv-12",
		"mcbd-25",
		"MKBD-S03",
		"FDD2002",
	})
}
