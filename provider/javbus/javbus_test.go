package javbus

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

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
		// Detail page exists, but missing from JavBus search index.
		// https://github.com/metatube-community/metatube-sdk-go/issues/381
		"RBD-859",
		// Missing the dash between label and number.
		"ssis277",
	})
}

func TestJavBus_Normalize(t *testing.T) {
	bus := New()
	for keyword, want := range map[string]string{
		"ssis277":     "SSIS-277",
		"MIDV005":     "MIDV-005",
		"SSIS-277":    "SSIS-277",
		"ssis-033":    "SSIS-033",
		"083120_001":  "083120_001",
		"heyzo-1234":  "HEYZO-1234",
		"FC2-1234567": "", // JavBus has no FC2 contents.
	} {
		require.Equal(t, want, bus.NormalizeMovieKeyword(keyword))
	}
	for id, want := range map[string]string{
		"ssis277":             "SSIS-277",
		"SSIS-277":            "SSIS-277",
		"MIDV-005_2025-02-04": "MIDV-005_2025-02-04",
		"n1234":               "N1234", // Single-letter label, keep as is.
	} {
		require.Equal(t, want, bus.NormalizeMovieID(id))
	}
}

func TestJavBus_SearchMovie_NotFound(t *testing.T) {
	if ci, _ := strconv.ParseBool(os.Getenv("GITHUB_ACTIONS")); ci {
		t.SkipNow()
	}
	bus := New()
	results, err := bus.SearchMovie(bus.NormalizeMovieKeyword("ZZZZ-9999"))
	require.NoError(t, err)
	require.Empty(t, results)
}
