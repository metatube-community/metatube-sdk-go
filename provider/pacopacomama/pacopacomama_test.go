package pacopacomama

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestPacopacomama_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"032622_623",
		"082107_257",
	})
}

func TestPacopacomama_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"032622_623",
	})
}
