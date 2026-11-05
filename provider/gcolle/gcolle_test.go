package gcolle

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestGcolle_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		// The site was revamped to /adult/contents/{newID}; the IDs below are
		// real, currently-existing ones verified in practice.
		// Old IDs (e.g. 847256) now 301-redirect to the takedown page via the
		// old path and are no longer usable.
		"55720",
		"55721",
		"55722",
	})
}
