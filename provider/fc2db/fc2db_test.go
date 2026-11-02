package fc2db

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFC2DB_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"4573362",
		"4922101",
		"4974437",
		"1192398",
		"4988735",
	})
}

func TestFC2DB_SearchMovie(t *testing.T) {
	p := New()
	results, err := p.SearchMovie("あまね")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("no results returned")
	}
	for _, r := range results {
		if r.ID == "" || r.Title == "" || r.Provider != Name {
			t.Errorf("incomplete result: %+v", r)
		}
	}
}
