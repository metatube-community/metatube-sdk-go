package sod

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestSOD_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"3DSVR-0416",
		"DLDSS-077",
		"3DSVR-1439",
		"START-114-V",
	})
}

func TestSOD_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"STAR-399",
		"IENF-209",
		"DLDSS-02",
		"START-114",
	})
}

func TestSOD_Fetch(t *testing.T) {
	testkit.Test(t, New, []string{
		"https://dy43ylo5q3vt8.cloudfront.net/_pics/202108/dldss_022/dldss_022_m.jpg",
	})
}
