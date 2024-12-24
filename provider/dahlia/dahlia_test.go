package dahlia

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestDAHLIA_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"dldss339",
		"DLDSS327",
	})
}

func TestDAHLIA_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"dldss-287",
		"DLDSS-259",
		"dldss271",
		"DLDSS274",
		"dldss087",
	})
}
