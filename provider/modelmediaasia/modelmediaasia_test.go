package modelmediaasia

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestModelMediaAsia_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"MTVQ18-EP1-AV",
		"MDCM-0013",
		"mdcm-0015",
	})
}

func TestModelMediaAsia_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"mdcm",
		"补习班",
	})
}

func TestModelMediaAsia_ParseMovieIDFromURL(t *testing.T) {
	provider := New()
	rawURL := "https://api.modelmediaasia.com/api/v2/videos/MDCM-0013"
	wantID := "MDCM-0013"
	gotID, err := provider.ParseMovieIDFromURL(rawURL)
	assert.NoError(t, err)
	assert.Equal(t, wantID, gotID)
}
