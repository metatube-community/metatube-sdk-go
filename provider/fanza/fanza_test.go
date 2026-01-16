package fanza

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestFANZA_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"prst00022",
		"196glod00359",
		"smub037",
		"1silk00113",
		"adn00306",
		"okax841",
		"zuko00122",
		"118chn064",
		"midv00047",
		"403jdxa57676",
		"pkpd00170",
		"mism00238",
		"1msfh00007",
		"1stars00141",
		"118abp906",
		"196glod0325t",
		"403jdxa057424",
		"xvsr00845",
	},
		testkit.FieldsNotEmpty("preview_images"),
		testkit.FieldsNotEmptyAny("maker", "label", "series"),
		testkit.FieldsNotEmptyAny("preview_video_url", "preview_video_hls_url"),
	)
}

func TestFANZA_GetMovieInfoByURL(t *testing.T) {
	t.Run("Legacy/Mono", func(t *testing.T) {
		testkit.Test(t, New, []string{
			"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=1start00190/",
			"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=mdvr00346/",
			"https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=neki004/",
			"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=mdvr00344/",
			"https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=k9sone740/",
			"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=ebwh00164/",
			"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=fpre00106/",
			"https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=41hodv21810/",
			"https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=h_346rebd655/",
			"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=ipvr00231/",
			"https://www.dmm.co.jp/mono/anime/-/detail/=/cid=196glod0323t/",
			"https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=fuyu079/",
		},
			testkit.FieldsNotEmpty("preview_images", "actors", "genres"),
			testkit.FieldsNotEmptyAny("maker", "label", "series"),
			testkit.FieldsNotEmptyAny("preview_video_url", "preview_video_hls_url"),
		)
	})

	t.Run("Video:AV/VR", func(t *testing.T) {
		testkit.Test(t, New, []string{
			"https://video.dmm.co.jp/av/content/?id=jux00109",
			"https://video.dmm.co.jp/av/content/?id=cawd00864",
			"https://video.dmm.co.jp/av/content/?id=juvr00240",
			"https://video.dmm.co.jp/av/content/?id=sivr00361",
			"https://video.dmm.co.jp/av/content/?id=fcvr00042",
			"https://video.dmm.co.jp/av/content/?id=1stars00273",
			"https://video.dmm.co.jp/av/content/?id=wanz00360",
		},
			testkit.FieldsNotEmpty("preview_images", "actors", "genres"),
			testkit.FieldsNotEmptyAny("maker", "label", "series"),
			testkit.FieldsNotEmptyAny("preview_video_url", "preview_video_hls_url"),
		)
	})

	t.Run("Video:Anime/Amateur", func(t *testing.T) {
		testkit.Test(t, New, []string{
			"https://video.dmm.co.jp/anime/content/?id=196glod00359",
			"https://video.dmm.co.jp/amateur/content/?id=smjx065",
		},
			testkit.FieldsNotEmpty("preview_images", "genres"),
			testkit.FieldsNotEmptyAny("maker", "label", "series"),
			testkit.FieldsNotEmptyAny("preview_video_url", "preview_video_hls_url"),
		)
	})

	t.Run("Video:Cinema", func(t *testing.T) {
		testkit.Test(t, New, []string{
			"https://video.dmm.co.jp/cinema/content/?id=5421ksd00051",
			"https://video.dmm.co.jp/cinema/content/?id=141nkt086",
		},
			testkit.FieldsNotEmpty("actors", "genres"),
			testkit.FieldsNotEmptyAny("maker", "label", "series"),
		)
	})

	t.Run("BigThumbURL", func(t *testing.T) {
		testkit.Test(t, New, []string{
			"https://video.dmm.co.jp/av/content/?id=sivr00361",
			"https://video.dmm.co.jp/av/content/?id=fcvr00042",
			"https://video.dmm.co.jp/av/content/?id=1stars00273",
			"https://video.dmm.co.jp/av/content/?id=13dsvr00157",
			"https://video.dmm.co.jp/av/content/?id=13dsvr00211",
		},
			testkit.FieldsNotEmpty("big_thumb_url", "preview_images"),
			testkit.FieldsNotEmptyAny("maker", "label", "series"),
		)
	})
}

func TestFANZA_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"PRST-022",
		"OREC-062",
		"T-28621",
		"midv-003",
		"blk-052",
		"mogi044",
		"SSNI-379",
		"SSIS-122",
		"MIDV-047",
		"abw",
		"HMN-686",
		"START-285",
	})
}

func TestFANZA_GetMovieReviewsByURL(t *testing.T) {
	testkit.Test(t, New, []string{
		"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=dass00256/",
		"https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=1fsdss301/",
		"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=ssis00964/",
		"https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=smus029/",
		"https://www.dmm.co.jp/digital/nikkatsu/-/detail/=/cid=5421ksd00051/",
		"https://www.dmm.co.jp/digital/anime/-/detail/=/cid=h_402mjad00329/",
		"https://www.dmm.co.jp/mono/anime/-/detail/=/cid=196glod0154/",
	})
}

func TestParseNumber(t *testing.T) {
	for _, unit := range []struct {
		id, want string
	}{
		{"ssis00123", "SSIS-123"},
		{"48midv00123", "MIDV-123"},
		{"48midv00003", "MIDV-003"},
		{"24ped00020", "PED-020"},
		{"abc00120", "ABC-120"},
		{"abc00120l", "ABC-120"},
		{"19abc00120l", "ABC-120"},
		{"abc00001", "ABC-001"},
		{"h_001fcp00006", "FCP-006"},
		{"001fcp06", "FCP-006"},
		{"h_001fcp06", "FCP-006"},
		{"scute1192", "SCUTE-1192"},
		{"h_198need00094r18", "NEED-094"},
		{"1fsdss00131re01", "FSDSS-131"},
		{"h_068mxgs1184bod", "MXGS-1184"},
		{"h_093r1800258", "R-1800258"},
		{"55t28621tk", "T-28621"},
	} {
		assert.Equal(t, unit.want, ParseNumber(unit.id))
	}
}

func TestPreviewSrc(t *testing.T) {
	for _, unit := range []struct {
		src, want string
	}{
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990ps.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990pl.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/consumer_game/pppd00990/pppd00990js-1.jpg",
			"https://pics.dmm.co.jp/digital/consumer_game/pppd00990/pppd00990-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990js-1.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990jp-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990ts-1.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990tl-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990-1.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990jp-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990-23.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990jp-23.jpg",
		},
	} {
		assert.Equal(t, unit.want, PreviewSrc(unit.src))
	}
}
