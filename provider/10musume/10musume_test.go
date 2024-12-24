package tenmusume

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestTenMusume_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"042922_01",
		"041607_01",
		"010906_04",
		"120409_01",
	})
}

func TestTenMusume_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"042922_01",
	})
}
