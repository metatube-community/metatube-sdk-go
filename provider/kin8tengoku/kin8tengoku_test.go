package kin8tengoku

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestKIN8_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"3604",
		"3556",
		"3580",
		"3521",
		"3587",
		"1045",
		"3591",
		"3421",
		"3600",
		"2508",
		"1662",
	})
}
