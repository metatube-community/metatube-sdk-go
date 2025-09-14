package strike3

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestBlacked_SearchMovie(t *testing.T) {
	testkit.Test(t, NewBlacked, []string{
		"Blacked - 2018-12-12 - The Real Thing",
	})
}

func TestBlacked_GetMovieInfoByURL(t *testing.T) {
	testkit.Test(t, NewBlacked, []string{
		"https://www.blacked.com/videos/the-real-thing",
	})
}

func TestBlacked_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, NewBlacked, []string{
		"the-real-thing",
	})
}
