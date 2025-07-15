//go:build deprecated

package arzon

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestARZON_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"1291144",
		"1252925",
		"1624669",
	})
}

func TestARZON_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"STARS",
		"IENF-209",
		"DLDSS-02",
		"FNEO-061",
	})
}

func TestARZON_Fetch(t *testing.T) {
	testkit.Test(t, New, []string{
		"https://img.arzon.jp/image/1/1663/1663651L.jpg",
	})
}
