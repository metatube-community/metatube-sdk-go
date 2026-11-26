package avbase

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

// assertSearchHomepage checks that every search result carries the canonical
// prefixed homepage, i.e. /works/<prefix>:<number> when a prefix is present.
func assertSearchHomepage() testkit.ValidateFunc {
	return func(t *testing.T, a any) {
		results, ok := a.([]*model.MovieSearchResult)
		require.True(t, ok)
		for _, r := range results {
			assert.Equalf(t, fmt.Sprintf(movieURL, r.ID), r.Homepage,
				"homepage mismatch for id: %s", r.ID)
		}
	}
}

func TestAVBase_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"prestige:ABP-588",
		"tameike:MEYD-856",
		"SSIS-354",
		"glory:GVH-186",
		"giga:GHMT-072",
	})
}

func TestAVBase_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"ABP-588",
		"MEYD-856",
		"SSIS-354",
		"HMN",
		"GHMT-072",
	}, assertSearchHomepage())
}
