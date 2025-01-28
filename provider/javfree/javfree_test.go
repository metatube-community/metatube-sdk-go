package javfree

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestJAVFREE_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"243452-1151912",
		"402171-4608186",
	})
}

func TestJAVFREE_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"FC2-PPV-1151912",
	})
}
