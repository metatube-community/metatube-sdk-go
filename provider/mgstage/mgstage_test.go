package mgstage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMGStage_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		//"TAP-002",
		"SIRO-2219",
		"300MAAN-778",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestMGStage_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"ABP-177",
		"200GANA-2701",
	} {
		results, err := provider.SearchMovie(provider.NormalizeMovieKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}

func TestMGStage_GetMovieReviewsByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"200GANA-2701",
		"MAAN-930",
		"217MTVR-048",
		"300MIUM-973",
		"MAAN-929",
		"ABP-177",
	} {
		reviews, err := provider.GetMovieReviewsByID(item)
		data, _ := json.MarshalIndent(reviews, "", "\t")
		if assert.NoError(t, err) {
			for _, review := range reviews {
				assert.True(t, review.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
