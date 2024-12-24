package madouqu

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestMadouQu_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"mdx-0267",
		"91cm109",
		"ras-361",
		"md0190-1",
		"pmc-472",
		"xg19",
		"hkd38",
		"dd002",
	})
}

func TestMadouQu_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"mdx",
		"厨房",
	})
}

func TestExtractImgSrc(t *testing.T) {
	for _, unit := range []struct {
		url, want string
	}{
		{
			"https://sp-ao.shortpixel.ai/client/to_auto,q_lossless,ret_img,w_717,h_569/https://madouqu.com/wp-content/uploads/2023/03/1678787961-49cc48460b5afc7.jpg",
			"https://madouqu.com/wp-content/uploads/2023/03/1678787961-49cc48460b5afc7.jpg",
		},
		{
			"https://sp-ao.shortpixel.ai/client/to_auto,q_lossless,ret_img,w_717,h_569/http://madouqu.com/wp-content/uploads/2023/03/1678787961-49cc48460b5afc7.jpg",
			"http://madouqu.com/wp-content/uploads/2023/03/1678787961-49cc48460b5afc7.jpg",
		},
	} {
		assert.Equal(t, unit.want, ExtractImgSrc(unit.url))
	}
}
