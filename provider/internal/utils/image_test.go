package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimilarImage(t *testing.T) {
	for i, item := range []struct {
		imgUrl1, imgUrl2 string
	}{
		// MSFH-025
		{
			"https://pics.dmm.co.jp/digital/video/1msfh00025/1msfh00025ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1msfh00025/1msfh00025jp-1.jpg",
		},
		// MSFH-037
		{
			"https://pics.dmm.co.jp/digital/video/1msfh00037/1msfh00037ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1msfh00037/1msfh00037jp-1.jpg",
		},
		// STARS-154
		{
			"https://pics.dmm.co.jp/digital/video/1stars00154/1stars00154ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1stars00154/1stars00154jp-1.jpg",
		},
		// STARS-249
		{
			"https://pics.dmm.co.jp/digital/video/1stars00249/1stars00249ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1stars00249/1stars00249jp-1.jpg",
		},
		// STARS-309
		{
			"https://pics.dmm.co.jp/digital/video/1stars00309/1stars00309ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1stars00309/1stars00309jp-1.jpg",
		},
		// STARS-325
		{
			"https://pics.dmm.co.jp/digital/video/1stars00325/1stars00325ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1stars00325/1stars00325jp-1.jpg",
		},
		// STARS-330
		{
			"https://pics.dmm.co.jp/digital/video/1stars00330/1stars00330ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1stars00330/1stars00330jp-1.jpg",
		},
		// SDDE-625
		{
			"https://pics.dmm.co.jp/digital/video/1sdde00625/1sdde00625ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1sdde00625/1sdde00625jp-1.jpg",
		},
		// SDMF-011
		{
			"https://pics.dmm.co.jp/digital/video/1sdmf00011/1sdmf00011ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1sdmf00011/1sdmf00011jp-1.jpg",
		},
		//{
		//	"https://pics.dmm.co.jp/digital/video/1sdmf00011/1sdmf00011ps.jpg",
		//	"https://pics.dmm.co.jp/digital/video/1sdmf00011/1sdmf00011jp-2.jpg",
		//},
		// SDMF-018
		{
			"https://pics.dmm.co.jp/digital/video/1sdmf00018/1sdmf00018ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1sdmf00018/1sdmf00018jp-1.jpg",
		},
		// SDMF-023
		{
			"https://pics.dmm.co.jp/digital/video/1sdmf00023/1sdmf00023ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1sdmf00023/1sdmf00023jp-1.jpg",
		},
		// STARS-200
		{
			"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200ps.jpg",
			"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200jp-1.jpg",
		},
		//{
		//	"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200ps.jpg",
		//	"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200jp-2.jpg",
		//},
		//{
		//	"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200jp-1.jpg",
		//	"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200jp-2.jpg",
		//},
		//{
		//	"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200pl.jpg",
		//	"https://pics.dmm.co.jp/digital/video/1stars00200/1stars00200jp-1.jpg",
		//},
	} {
		v := SimilarImage(item.imgUrl1, item.imgUrl2, nil)
		if v {
			t.Logf("No. %d is similar.", i)
		} else {
			t.Logf("No. %d is distinct.", i)
		}
		assert.True(t, v)
	}
}
