package kin8tengoku

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestKIN8_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		// After the site was revamped into a Next.js app, the following IDs are
		// real, verified ones.
		"4057",
		"4052",
		"4015",
		"4003",
		"3995",
		"3605",
		"3489",
	})
}
