package madouqu

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMadouQu_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"mdx-0267",
		"91cm109",
		"ras-361",
		"md0190-1",
		"pmc-472",
		"xg19",
		"hkd38",
		"dd002",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestMadouQu_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"mdx",
		"厨房",
	} {
		results, err := provider.SearchMovie(provider.NormalizeMovieKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
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
