package caribbeancom

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaribbeancom_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"050422-001",
		"031222-001",
		"061014-618",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestCaribbeancom_GetMovieReviewsByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"050422-001",
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
