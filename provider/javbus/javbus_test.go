package javbus

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestJavBus_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"SMBD-77",
		"SSNI-776",
		"ABP-331",
		"CEMD-232",
		"th101-000-110942",
	})
}

func TestJavBus_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"SSIS-033",
		"MIDV-005",
	})
}
