package heyzo

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestHeyzo_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"0841",
		"0805",
		"2189",
	})
}

func TestHeyzo_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"2949",
		"0328",
		"0805",
	})
}
