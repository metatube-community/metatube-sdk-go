package jav321

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestJAV321_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"heyzo2818",
		"300maan-791",
		"sivr00215",
		"ebod00916",
		"118abp00559",
		"nima00011",
		"pred00402",
	})
}

func TestJAV321_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"SSIS-033",
		"MIDV-005",
	})
}
