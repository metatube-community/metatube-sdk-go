package javfree

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestJAVFREE_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"IPX-001",
	})
}

func TestJAVFREE_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"IPX-001",
	})
}
