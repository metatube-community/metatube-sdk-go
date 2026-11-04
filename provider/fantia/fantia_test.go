package fantia

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFantia_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"887620",
	})
}
