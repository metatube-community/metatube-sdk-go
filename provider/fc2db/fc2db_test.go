package fc2db

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFC2DB_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"4573362",
		"4922101",
		"4974437",
		"1192398",
		"4988735",
	})
}

func TestFC2DB_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"FC2-PPV-4573362",
		"fc2-4922101",
	})
}
