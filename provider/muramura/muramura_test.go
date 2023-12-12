package muramura

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMuraMura_NormalizeID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"091522_959",
		"062509_011",
		"021110_163",
		"013010_157",
		"012810_155",
		"081222_953",
		"062509_003",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestMuraMura_GetReviewInfo(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"091522_959",
	} {
		reviews, err := provider.GetMovieReviewInfo(item)
		data, _ := json.MarshalIndent(reviews, "", "\t")
		if assert.NoError(t, err) {
			for _, review := range reviews {
				assert.True(t, review.Valid())
			}
		}
		t.Logf("%s", data)
	}
}
