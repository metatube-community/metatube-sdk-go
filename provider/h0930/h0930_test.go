package h0930

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestH0930_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"ori1643",
		"ori1492",
		"ori1396",
		"orijuku823",
		"orimrs695",
	})
}
