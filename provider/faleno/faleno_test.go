package faleno

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFALENO_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"fsdss754",
		"FSDSS749",
		"fcdss072",
		"FCDSS060",
	})
}

func TestFALENO_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"FSDSS-723",
		"FSDSS746",
		"fsdss728",
		"fsdss-721",
		"FCDSS-069",
		"FCDSS066",
	})
}
