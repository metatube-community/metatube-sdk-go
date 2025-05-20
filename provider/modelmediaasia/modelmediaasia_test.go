package modelmediaasia

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestModelMediaAsia_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, NewMovieProvider, []string{
		"MTVQ18-EP1-AV",
		"MDCM-0013",
		"mdcm-0015",
	})
}

func TestModelMediaAsia_GetMovieInfoByURL(t *testing.T) {
	testkit.Test(t, NewMovieProvider, []string{
		fmt.Sprintf(movieURL, "MTVQ18-EP1-AV"),
		fmt.Sprintf(movieURL, "MDCM-0013"),
		fmt.Sprintf(movieURL, "mdcm-0015"),
	})
}

func TestModelMediaAsia_SearchMovie(t *testing.T) {
	testkit.Test(t, NewMovieProvider, []string{
		"mdcm",
		"补习班",
	})
}

func TestModelMediaAsia_ParseMovieIDFromURL(t *testing.T) {
	provider := NewMovieProvider()
	rawURL := "https://api.modelmediaasia.com/api/v2/videos/MDCM-0013"
	wantID := "MDCM-0013"
	gotID, err := provider.ParseMovieIDFromURL(rawURL)
	assert.NoError(t, err)
	assert.Equal(t, wantID, gotID)
}

func TestModelMediaAsiaActor_GetActorInfoByID(t *testing.T) {
	testkit.Test(t, NewActorProvider, []string{
		"9",
		"11",
		"15",
		"62",
		"115",
	})
}

func TestModelMediaAsiaActor_GetActorInfoByURL(t *testing.T) {
	testkit.Test(t, NewActorProvider, []string{
		fmt.Sprintf(actorURL, "11"),
		fmt.Sprintf(actorURL, "15"),
	})
}

func TestModelMediaAsiaActor_SearchActor(t *testing.T) {
	testkit.Test(t, NewActorProvider, []string{
		"夏",
		"赵",
	})
}

func TestModelMediaAsiaActor_ParseActorIDFromURL(t *testing.T) {
	provider := NewActorProvider()
	rawURL := "https://modelmediaasia.com/zh-CN/models/15"
	wantID := "15"
	gotID, err := provider.ParseActorIDFromURL(rawURL)
	assert.NoError(t, err)
	assert.Equal(t, wantID, gotID)
}
