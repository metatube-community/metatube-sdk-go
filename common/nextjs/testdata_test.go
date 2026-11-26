package nextjs

import (
	"embed"
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// Pages in testdata are base64-encoded and named by their MD5:
//
//	gbase64 page.html > "$(md5sum page.html | cut -d' ' -f1)"
//
//go:embed testdata/*
var testdata embed.FS

type TestdataTestSuite struct {
	suite.Suite

	payloads map[string]string // filename -> RSC payload
}

func TestTestdataTestSuite(t *testing.T) {
	suite.Run(t, new(TestdataTestSuite))
}

// SetupSuite extracts the RSC payload of every page in testdata.
func (s *TestdataTestSuite) SetupSuite() {
	entries, err := testdata.ReadDir("testdata")
	s.Require().NoError(err)

	s.payloads = make(map[string]string, len(entries))
	for _, entry := range entries {
		data, err := testdata.ReadFile("testdata/" + entry.Name())
		s.Require().NoError(err)
		page, err := base64.StdEncoding.DecodeString(string(data))
		s.Require().NoError(err, entry.Name())
		payload, err := ExtractPayload(page)
		s.Require().NoError(err, entry.Name())
		s.Require().NotEmpty(payload, entry.Name())
		s.payloads[entry.Name()] = payload
	}
}

func (s *TestdataTestSuite) TestFindObject() {
	for _, unit := range []struct {
		filename string
		want     any // struct value to find, matched by its fields.
	}{
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			want: struct {
				Src                      string `json:"src"`
				Type                     string `json:"type"`
				Poster                   string `json:"poster"`
				SeekPreviewVTTURL        string `json:"seekPreviewVttUrl"`
				MovieSeq                 int    `json:"movieSeq"`
				Locale                   string `json:"locale"`
				PlaybackMode             string `json:"playbackMode"`
				StreamHistoryCookieKey   string `json:"streamHistoryCookieKey"`
				StreamingPositionKey     string `json:"streamingPositionKey"`
				IsHeyzoMember            bool   `json:"isHeyzoMember"`
				IsHeyzoMembershipExpired bool   `json:"isHeyzoMembershipExpired"`
				SceneThumbnails          []struct {
					ThumbNum    int    `json:"thumbNum"`
					ImageURL    string `json:"imageUrl"`
					Comment     string `json:"comment"`
					SeekSeconds int    `json:"seekSeconds"`
				} `json:"sceneThumbnails"`
			}{
				Src:                      "https://hls.heyzo.com/sample/3000/3986/mb.m3u8",
				Type:                     "application/x-mpegURL",
				Poster:                   "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg",
				SeekPreviewVTTURL:        "",
				MovieSeq:                 4001,
				Locale:                   "ja",
				PlaybackMode:             "sample",
				StreamHistoryCookieKey:   "viewed_4001",
				StreamingPositionKey:     "streaming_position_4001",
				IsHeyzoMember:            false,
				IsHeyzoMembershipExpired: false,
				SceneThumbnails: []struct {
					ThumbNum    int    `json:"thumbNum"`
					ImageURL    string `json:"imageUrl"`
					Comment     string `json:"comment"`
					SeekSeconds int    `json:"seekSeconds"`
				}{
					{
						ThumbNum:    1,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture1.jpg",
						Comment:     "耳なめ",
						SeekSeconds: 592,
					},
					{
						ThumbNum:    2,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture2.jpg",
						Comment:     "手マン",
						SeekSeconds: 766,
					},
					{
						ThumbNum:    3,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture3.jpg",
						Comment:     "クンニ",
						SeekSeconds: 815,
					},
					{
						ThumbNum:    4,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture4.jpg",
						Comment:     "バック",
						SeekSeconds: 960,
					},
					{
						ThumbNum:    5,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture5.jpg",
						Comment:     "キス",
						SeekSeconds: 1338,
					},
					{
						ThumbNum:    6,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture6.jpg",
						Comment:     "電マ",
						SeekSeconds: 1735,
					},
					{
						ThumbNum:    7,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture7.jpg",
						Comment:     "フェラ",
						SeekSeconds: 1880,
					},
					{
						ThumbNum:    8,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture8.jpg",
						Comment:     "69",
						SeekSeconds: 2340,
					},
					{
						ThumbNum:    9,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture9.jpg",
						Comment:     "正常位",
						SeekSeconds: 2470,
					},
					{
						ThumbNum:    10,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture10.jpg",
						Comment:     "騎乗後背位",
						SeekSeconds: 2802,
					},
					{
						ThumbNum:    11,
						ImageURL:    "https://www.heyzo.com/contents/3000/3986/images/capture11.jpg",
						Comment:     "中出し",
						SeekSeconds: 3485,
					},
				},
			},
		},
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			want: struct {
				MovieSeq         int       `json:"movie_seq"`
				MovieID          string    `json:"movie_id"`
				MonthlyStartDate time.Time `json:"monthly_start_date"`
				MonthlyEndDate   time.Time `json:"monthly_end_date"`
				ActorID          int       `json:"actor_id"`
				Name             string    `json:"name"`
				NameEN           string    `json:"name_en"`
				CreateDate       time.Time `json:"create_date"`
				UpdateDate       time.Time `json:"update_date"`
				Detail           struct {
					MovieSeq        int     `json:"movie_seq"`
					Memo            string  `json:"memo"`
					RvScore         float64 `json:"rv_score"`
					ImportedFromURL string  `json:"imported_from_url"`
					GroupID         int     `json:"group_id"`
				} `json:"detail"`
				Actors     []string `json:"actors"`
				Categories []struct {
					CategoryID        int    `json:"category_id"`
					CategoryName      string `json:"category_name"`
					CategoryNameEN    string `json:"category_name_en"`
					CategoryKeyName   string `json:"category_key_name"`
					Level             int    `json:"LEVEL"`
					DescriptionText   string `json:"description_text"`
					DescriptionTextEN string `json:"description_text_en"`
				} `json:"categories"`
				Download struct {
					Whole []struct {
						MovieSeq      int    `json:"movie_seq"`
						FileNum       int    `json:"file_num"`
						FileSubNumber int    `json:"file_sub_number"`
						Duration      string `json:"duration"`
						FileSize      int    `json:"file_size"`
						FileSizeLight int    `json:"file_size_light"`
						FileSizeMb    int    `json:"file_size_mb"`
						FiletypeID    int    `json:"filetype_id"`
					} `json:"whole"`
					Split []any `json:"split"`
				} `json:"download"`
				Attributes []struct {
					MovieSeq       int    `json:"movie_seq"`
					AttributeID    int    `json:"attribute_id"`
					AttributeValue string `json:"attribute_value"`
				} `json:"attributes"`
				MovieThumbnails []struct {
					ThumbNum int       `json:"thumb_num"`
					Time     time.Time `json:"time"`
					Comment  string    `json:"comment"`
				} `json:"movie_thumbnails"`
				ActorType          string `json:"actorType"`
				IsPreRelease       bool   `json:"isPreRelease"`
				AvailableDays      int    `json:"availableDays"`
				IsLimited          bool   `json:"isLimited"`
				IsBonus            bool   `json:"isBonus"`
				TotalViewers       int    `json:"totalViewers"`
				ReviewCommentCount int    `json:"reviewCommentCount"`
				MovieGrade         string `json:"movieGrade"`
				MonthlyStatus      bool   `json:"monthly_status"`
				MovieFiles         []struct {
					MovieSeq      int       `json:"movie_seq"`
					FileNum       int       `json:"file_num"`
					FileSubNumber int       `json:"file_sub_number"`
					Duration      time.Time `json:"duration"`
					FileSize      int       `json:"file_size"`
					FileSizeLight int       `json:"file_size_light"`
					FileSizeMb    int       `json:"file_size_mb"`
					FiletypeID    int       `json:"filetype_id"`
				} `json:"movie_files"`
				MovieStreamAggregate struct {
					TotalViewers int `json:"total_viewers"`
				} `json:"movie_stream_aggregate"`
				Count struct {
					Reviews int `json:"reviews"`
				} `json:"_count"`
			}{
				MovieSeq:         4001,
				MovieID:          "3986",
				MonthlyStartDate: time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC),
				MonthlyEndDate:   time.Date(2100, 12, 31, 0, 0, 0, 0, time.UTC),
				ActorID:          1705,
				Name:             "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2",
				NameEN:           "Endure Male Pornstars'S Caresses, Or Get Something Vol.2",
				CreateDate:       time.Date(2026, 8, 31, 22, 56, 25, 0, time.UTC),
				UpdateDate:       time.Date(2026, 10, 22, 16, 22, 47, 0, time.UTC),
				Detail: struct {
					MovieSeq        int     `json:"movie_seq"`
					Memo            string  `json:"memo"`
					RvScore         float64 `json:"rv_score"`
					ImportedFromURL string  `json:"imported_from_url"`
					GroupID         int     `json:"group_id"`
				}{
					MovieSeq:        4001,
					Memo:            "前回の出演作が好評だった羽柴よしのちゃんが再びHEYZOに登場！前回が楽しかったので今回も楽しみというよしのちゃん、今回は企画の内容を知らされないまま来ちゃいましたが、AV男優の凄テクに10分間我慢できたらボーナスタイムがもらえるという企画！しかし、我慢できずに罰ゲームを受けることになってしまいました・・・・でもせっかく来たのでボーナスタイムも一緒に楽しんでいくことになって、最後は中出しフィニッシュで大満足のよしのちゃんでした。",
					RvScore:         0,
					ImportedFromURL: "",
					GroupID:         2,
				},
				Actors: []string{
					"$6:props:children:1:props:children:0:props:actor",
				},
				Categories: []struct {
					CategoryID        int    `json:"category_id"`
					CategoryName      string `json:"category_name"`
					CategoryNameEN    string `json:"category_name_en"`
					CategoryKeyName   string `json:"category_key_name"`
					Level             int    `json:"LEVEL"`
					DescriptionText   string `json:"description_text"`
					DescriptionTextEN string `json:"description_text_en"`
				}{
					{
						CategoryID:        1,
						CategoryName:      "69",
						CategoryNameEN:    "69",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        3,
						CategoryName:      "中出し",
						CategoryNameEN:    "Creampie",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        13,
						CategoryName:      "おもちゃ",
						CategoryNameEN:    "Toy",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        15,
						CategoryName:      "騎乗位",
						CategoryNameEN:    "Riding",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        16,
						CategoryName:      "バック",
						CategoryNameEN:    "Doggy Style",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        20,
						CategoryName:      "AV女優",
						CategoryNameEN:    "Porn Star",
						CategoryKeyName:   "",
						Level:             2,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        29,
						CategoryName:      "美乳",
						CategoryNameEN:    "Nice Tits",
						CategoryKeyName:   "",
						Level:             2,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        31,
						CategoryName:      "ギャル",
						CategoryNameEN:    "Gal",
						CategoryKeyName:   "",
						Level:             2,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        39,
						CategoryName:      "予告作品フラグ",
						CategoryNameEN:    "release preview",
						CategoryKeyName:   "op_flag_release_preview",
						Level:             3,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        157,
						CategoryName:      "その他",
						CategoryNameEN:    "Others",
						CategoryKeyName:   "",
						Level:             7,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        178,
						CategoryName:      "パイパン",
						CategoryNameEN:    "Glabrousness",
						CategoryKeyName:   "",
						Level:             2,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        280,
						CategoryName:      "指マン",
						CategoryNameEN:    "Finger Fuck",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        281,
						CategoryName:      "クンニ",
						CategoryNameEN:    "Cunnilingus",
						CategoryKeyName:   "",
						Level:             1,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
					{
						CategoryID:        317,
						CategoryName:      "自社２次利用作品",
						CategoryNameEN:    "No Heydouga export",
						CategoryKeyName:   "op_flag_inhouse_reuse",
						Level:             3,
						DescriptionText:   "",
						DescriptionTextEN: "",
					},
				},
				Download: struct {
					Whole []struct {
						MovieSeq      int    `json:"movie_seq"`
						FileNum       int    `json:"file_num"`
						FileSubNumber int    `json:"file_sub_number"`
						Duration      string `json:"duration"`
						FileSize      int    `json:"file_size"`
						FileSizeLight int    `json:"file_size_light"`
						FileSizeMb    int    `json:"file_size_mb"`
						FiletypeID    int    `json:"filetype_id"`
					} `json:"whole"`
					Split []any `json:"split"`
				}{
					Whole: []struct {
						MovieSeq      int    `json:"movie_seq"`
						FileNum       int    `json:"file_num"`
						FileSubNumber int    `json:"file_sub_number"`
						Duration      string `json:"duration"`
						FileSize      int    `json:"file_size"`
						FileSizeLight int    `json:"file_size_light"`
						FileSizeMb    int    `json:"file_size_mb"`
						FiletypeID    int    `json:"filetype_id"`
					}{
						{
							MovieSeq:      4001,
							FileNum:       1,
							FileSubNumber: 1,
							Duration:      "00:59:48",
							FileSize:      2184,
							FileSizeLight: 901,
							FileSizeMb:    216,
							FiletypeID:    1,
						},
					},
					Split: []any{},
				},
				Attributes: []struct {
					MovieSeq       int    `json:"movie_seq"`
					AttributeID    int    `json:"attribute_id"`
					AttributeValue string `json:"attribute_value"`
				}{
					{
						MovieSeq:       4001,
						AttributeID:    1,
						AttributeValue: "金髪",
					},
					{
						MovieSeq:       4001,
						AttributeID:    5,
						AttributeValue: "0",
					},
					{
						MovieSeq:       4001,
						AttributeID:    9,
						AttributeValue: "",
					},
					{
						MovieSeq:       4001,
						AttributeID:    19,
						AttributeValue: "{\"full\":\"00:59:48\"}",
					},
				},
				MovieThumbnails: []struct {
					ThumbNum int       `json:"thumb_num"`
					Time     time.Time `json:"time"`
					Comment  string    `json:"comment"`
				}{
					{
						ThumbNum: 1,
						Time:     time.Date(1970, 1, 1, 0, 9, 52, 0, time.UTC),
						Comment:  "耳なめ",
					},
					{
						ThumbNum: 2,
						Time:     time.Date(1970, 1, 1, 0, 12, 46, 0, time.UTC),
						Comment:  "手マン",
					},
					{
						ThumbNum: 3,
						Time:     time.Date(1970, 1, 1, 0, 13, 35, 0, time.UTC),
						Comment:  "クンニ",
					},
					{
						ThumbNum: 4,
						Time:     time.Date(1970, 1, 1, 0, 16, 0, 0, time.UTC),
						Comment:  "バック",
					},
					{
						ThumbNum: 5,
						Time:     time.Date(1970, 1, 1, 0, 22, 18, 0, time.UTC),
						Comment:  "キス",
					},
					{
						ThumbNum: 6,
						Time:     time.Date(1970, 1, 1, 0, 28, 55, 0, time.UTC),
						Comment:  "電マ",
					},
					{
						ThumbNum: 7,
						Time:     time.Date(1970, 1, 1, 0, 31, 20, 0, time.UTC),
						Comment:  "フェラ",
					},
					{
						ThumbNum: 8,
						Time:     time.Date(1970, 1, 1, 0, 39, 0, 0, time.UTC),
						Comment:  "69",
					},
					{
						ThumbNum: 9,
						Time:     time.Date(1970, 1, 1, 0, 41, 10, 0, time.UTC),
						Comment:  "正常位",
					},
					{
						ThumbNum: 10,
						Time:     time.Date(1970, 1, 1, 0, 46, 42, 0, time.UTC),
						Comment:  "騎乗後背位",
					},
					{
						ThumbNum: 11,
						Time:     time.Date(1970, 1, 1, 0, 58, 5, 0, time.UTC),
						Comment:  "中出し",
					},
				},
				ActorType:          "av",
				IsPreRelease:       false,
				AvailableDays:      27064,
				IsLimited:          false,
				IsBonus:            false,
				TotalViewers:       74,
				ReviewCommentCount: 0,
				MovieGrade:         "normal",
				MonthlyStatus:      true,
				MovieFiles: []struct {
					MovieSeq      int       `json:"movie_seq"`
					FileNum       int       `json:"file_num"`
					FileSubNumber int       `json:"file_sub_number"`
					Duration      time.Time `json:"duration"`
					FileSize      int       `json:"file_size"`
					FileSizeLight int       `json:"file_size_light"`
					FileSizeMb    int       `json:"file_size_mb"`
					FiletypeID    int       `json:"filetype_id"`
				}{
					{
						MovieSeq:      4001,
						FileNum:       1,
						FileSubNumber: 1,
						Duration:      time.Date(1970, 1, 1, 0, 59, 48, 0, time.UTC),
						FileSize:      2184,
						FileSizeLight: 901,
						FileSizeMb:    216,
						FiletypeID:    1,
					},
				},
				MovieStreamAggregate: struct {
					TotalViewers int `json:"total_viewers"`
				}{
					TotalViewers: 74,
				},
				Count: struct {
					Reviews int `json:"reviews"`
				}{
					Reviews: 0,
				},
			},
		},
		{
			// https://www.dmm.co.jp/search/=/searchstr=SSIS-419/limit=30/sort=rankprofile
			filename: "a0ef9a6847f4d22cca6abc9480b7a3ae",
			want: struct {
				SearchWord string `json:"search_word"`
				Limit      int    `json:"limit"`
				Page       int    `json:"page"`
				OrderBy    string `json:"order_by"`
				Contents   struct {
					Count            int  `json:"count"`
					ExceededHitLimit bool `json:"exceeded_hit_limit"`
					Data             []struct {
						UID                 string   `json:"uid"`
						ContentID           string   `json:"content_id"`
						Floor               string   `json:"floor"`
						ServiceName         string   `json:"service_name"`
						SearchServiceURL    string   `json:"search_service_url"`
						Title               string   `json:"title"`
						DetailURL           string   `json:"detail_url"`
						CurrentPrice        int      `json:"current_price"`
						ItemPricing         string   `json:"item_pricing"`
						ThumbnailImageURL   string   `json:"thumbnail_image_url"`
						Rate                float64  `json:"rate"`
						ReviewCount         int      `json:"review_count"`
						MediaType           string   `json:"media_type"`
						ReleaseStatuses     []string `json:"release_statuses"`
						ReleaseAnnouncement string   `json:"release_announcement"`
						Keywords            []string `json:"keywords"`
						Casts               []string `json:"casts"`
						Authors             []string `json:"authors"`
						Makers              []string `json:"makers"`
						Series              string   `json:"series"`
						Comment             string   `json:"comment"`
						SaleStatus          string   `json:"sale_status"`
						OriginalPrice       int      `json:"original_price"`
						DiscountRate        int      `json:"discount_rate"`
						SampleMovieURL      string   `json:"sample_movie_url"`
						IsAvailable         bool     `json:"is_available"`
					} `json:"data"`
				} `json:"contents"`
				Facets struct {
					Services []struct {
						Value string `json:"value"`
						Label string `json:"label"`
						URL   string `json:"url"`
					} `json:"services"`
					Categories []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					} `json:"categories"`
					ReleaseDates []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					} `json:"release_dates"`
					Reviews []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					} `json:"reviews"`
				} `json:"facets"`
				Pagination struct {
					CurrentPage int `json:"currentPage"`
					PagesCount  int `json:"pagesCount"`
				} `json:"pagination"`
				CategoryLinks        []any  `json:"category_links"`
				AbTestGroup          string `json:"ab_test_group"`
				FilterFeatureEnabled bool   `json:"filter_feature_enabled"`
			}{
				SearchWord: "SSIS-419",
				Limit:      30,
				Page:       1,
				OrderBy:    "rankprofile",
				Contents: struct {
					Count            int  `json:"count"`
					ExceededHitLimit bool `json:"exceeded_hit_limit"`
					Data             []struct {
						UID                 string   `json:"uid"`
						ContentID           string   `json:"content_id"`
						Floor               string   `json:"floor"`
						ServiceName         string   `json:"service_name"`
						SearchServiceURL    string   `json:"search_service_url"`
						Title               string   `json:"title"`
						DetailURL           string   `json:"detail_url"`
						CurrentPrice        int      `json:"current_price"`
						ItemPricing         string   `json:"item_pricing"`
						ThumbnailImageURL   string   `json:"thumbnail_image_url"`
						Rate                float64  `json:"rate"`
						ReviewCount         int      `json:"review_count"`
						MediaType           string   `json:"media_type"`
						ReleaseStatuses     []string `json:"release_statuses"`
						ReleaseAnnouncement string   `json:"release_announcement"`
						Keywords            []string `json:"keywords"`
						Casts               []string `json:"casts"`
						Authors             []string `json:"authors"`
						Makers              []string `json:"makers"`
						Series              string   `json:"series"`
						Comment             string   `json:"comment"`
						SaleStatus          string   `json:"sale_status"`
						OriginalPrice       int      `json:"original_price"`
						DiscountRate        int      `json:"discount_rate"`
						SampleMovieURL      string   `json:"sample_movie_url"`
						IsAvailable         bool     `json:"is_available"`
					} `json:"data"`
				}{
					Count:            3,
					ExceededHitLimit: false,
					Data: []struct {
						UID                 string   `json:"uid"`
						ContentID           string   `json:"content_id"`
						Floor               string   `json:"floor"`
						ServiceName         string   `json:"service_name"`
						SearchServiceURL    string   `json:"search_service_url"`
						Title               string   `json:"title"`
						DetailURL           string   `json:"detail_url"`
						CurrentPrice        int      `json:"current_price"`
						ItemPricing         string   `json:"item_pricing"`
						ThumbnailImageURL   string   `json:"thumbnail_image_url"`
						Rate                float64  `json:"rate"`
						ReviewCount         int      `json:"review_count"`
						MediaType           string   `json:"media_type"`
						ReleaseStatuses     []string `json:"release_statuses"`
						ReleaseAnnouncement string   `json:"release_announcement"`
						Keywords            []string `json:"keywords"`
						Casts               []string `json:"casts"`
						Authors             []string `json:"authors"`
						Makers              []string `json:"makers"`
						Series              string   `json:"series"`
						Comment             string   `json:"comment"`
						SaleStatus          string   `json:"sale_status"`
						OriginalPrice       int      `json:"original_price"`
						DiscountRate        int      `json:"discount_rate"`
						SampleMovieURL      string   `json:"sample_movie_url"`
						IsAvailable         bool     `json:"is_available"`
					}{
						{
							UID:                 "bb5d76de9509ff38c25992920aae855b",
							ContentID:           "ssis419",
							Floor:               "mono_dvd",
							ServiceName:         "通販",
							SearchServiceURL:    "https://www.dmm.co.jp/mono/-/search/=/searchstr=SSIS-419/",
							Title:               "初体験は三上悠亜 童貞喪失してあげたその日からひたすらセックスに明け暮れた2人きりの筆おろしドリーム同棲ドキュメント",
							DetailURL:           "https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=ssis419/?i3_ref=search&i3_ord=1",
							CurrentPrice:        3502,
							ItemPricing:         "single",
							ThumbnailImageURL:   "https://pics.dmm.co.jp/mono/movie/adult/ssis419/ssis419ps.jpg",
							Rate:                3,
							ReviewCount:         29,
							MediaType:           "DVD",
							ReleaseStatuses:     []string{},
							ReleaseAnnouncement: "発売日：2022/6/14",
							Keywords: []string{
								"サンプル動画",
								"単体作品",
								"童貞",
								"アイドル・芸能人",
								"ドキュメンタリー",
								"ギリモザ",
							},
							Casts: []string{
								"三上悠亜",
							},
							Authors: nil,
							Makers: []string{
								"エスワン ナンバーワンスタイル",
							},
							Series:         "筆おろし同棲ドキュメント",
							Comment:        "",
							SaleStatus:     "%OFF",
							OriginalPrice:  3502,
							DiscountRate:   0,
							SampleMovieURL: "https://tencc3001.dmm.co.jp/pv/TQgoWApW4Oa-Y28hRZux3j47IJHTm_c7efTUMsC9XpmEncdmsWABRwgYOFZ2pt/ssis00419_sm_w.mp4",
							IsAvailable:    false,
						},
						{
							UID:                 "afd959929ed28c976b8efb63b52c5db3",
							ContentID:           "9ssis419",
							Floor:               "mono_dvd",
							ServiceName:         "通販",
							SearchServiceURL:    "https://www.dmm.co.jp/mono/-/search/=/searchstr=SSIS-419/",
							Title:               "初体験は三上悠亜 童貞喪失してあげたその日からひたすらセックスに明け暮れた2人きりの筆おろしドリーム同棲ドキュメント （ブルーレイディスク）",
							DetailURL:           "https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=9ssis419/?i3_ref=search&i3_ord=2",
							CurrentPrice:        3834,
							ItemPricing:         "single",
							ThumbnailImageURL:   "https://pics.dmm.co.jp/mono/movie/adult/9ssis419/9ssis419ps.jpg",
							Rate:                4,
							ReviewCount:         2,
							MediaType:           "Blu-ray",
							ReleaseStatuses:     []string{},
							ReleaseAnnouncement: "発売日：2022/6/14",
							Keywords: []string{
								"セール対象商品",
								"エスワンキャンペーン",
								"サンプル動画",
								"Blu-ray（ブルーレイ）",
								"単体作品",
								"童貞",
								"アイドル・芸能人",
								"ドキュメンタリー",
							},
							Casts: []string{
								"三上悠亜",
							},
							Authors: nil,
							Makers: []string{
								"エスワン ナンバーワンスタイル",
							},
							Series:         "筆おろし同棲ドキュメント",
							Comment:        "",
							SaleStatus:     "%OFF",
							OriginalPrice:  3834,
							DiscountRate:   30,
							SampleMovieURL: "https://tencc3001.dmm.co.jp/pv/TQgoWApW4Oa-Y28hRZux3j47IJHTm_c7efTUMsC9XpmEncdmsWABRwgYOFZ2pt/ssis00419_sm_w.mp4",
							IsAvailable:    true,
						},
						{
							UID:               "8e97b490a3fd2d4ff2ce98e17ecb4470",
							ContentID:         "4ssis419",
							Floor:             "rental_dvd",
							ServiceName:       "DVDレンタル",
							SearchServiceURL:  "https://www.dmm.co.jp/rental/-/search/=/searchstr=SSIS-419/",
							Title:             "初体験は三上悠亜 童貞喪失してあげたその日からひたすらセックスに明け暮れた2人きりの筆おろしドリーム同棲ドキュメント",
							DetailURL:         "https://www.dmm.co.jp/rental/ppr/-/detail/=/cid=4ssis419/?i3_ref=search&i3_ord=3",
							CurrentPrice:      299,
							ItemPricing:       "multi",
							ThumbnailImageURL: "https://pics.dmm.co.jp/mono/movie/4ssis419/4ssis419ps.jpg",
							Rate:              3,
							ReviewCount:       29,
							MediaType:         "DVD",
							ReleaseStatuses: []string{
								"旧作",
								"独占",
							},
							ReleaseAnnouncement: "貸出日：2022/9/24",
							Keywords: []string{
								"サンプル動画",
								"単体作品",
								"童貞",
								"ギリモザ",
								"アイドル・芸能人",
								"ドキュメンタリー",
								"セル仕様",
								"独占レンタル",
							},
							Casts: []string{
								"三上悠亜",
							},
							Authors: nil,
							Makers: []string{
								"エスワン ナンバーワンスタイル",
							},
							Series:         "",
							Comment:        "",
							SaleStatus:     "キャンペーンなし",
							OriginalPrice:  598,
							DiscountRate:   0,
							SampleMovieURL: "",
							IsAvailable:    true,
						},
					},
				},
				Facets: struct {
					Services []struct {
						Value string `json:"value"`
						Label string `json:"label"`
						URL   string `json:"url"`
					} `json:"services"`
					Categories []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					} `json:"categories"`
					ReleaseDates []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					} `json:"release_dates"`
					Reviews []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					} `json:"reviews"`
				}{
					Services: []struct {
						Value string `json:"value"`
						Label string `json:"label"`
						URL   string `json:"url"`
					}{
						{
							Value: "rental",
							Label: "DVDレンタル",
							URL:   "https://www.dmm.co.jp/rental/-/search/=/searchstr=SSIS-419/",
						},
						{
							Value: "mono",
							Label: "通販",
							URL:   "https://www.dmm.co.jp/mono/-/search/=/searchstr=SSIS-419/",
						},
					},
					Categories: []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					}{
						{
							Value: "video_purchase",
							Count: 2,
						},
						{
							Value: "video_monthly",
							Count: 0,
						},
						{
							Value: "video_rental",
							Count: 1,
						},
						{
							Value: "book_purchase",
							Count: 0,
						},
						{
							Value: "book_monthly",
							Count: 0,
						},
						{
							Value: "doujin",
							Count: 0,
						},
						{
							Value: "game",
							Count: 0,
						},
						{
							Value: "goods",
							Count: 0,
						},
					},
					ReleaseDates: []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					}{
						{
							Value: "within_3_days",
							Count: 0,
						},
						{
							Value: "within_7_days",
							Count: 0,
						},
						{
							Value: "within_1_month",
							Count: 0,
						},
						{
							Value: "upcoming",
							Count: 0,
						},
					},
					Reviews: []struct {
						Value string `json:"value"`
						Count int    `json:"count"`
					}{
						{
							Value: "4.5",
							Count: 0,
						},
						{
							Value: "4.0",
							Count: 1,
						},
						{
							Value: "3.5",
							Count: 1,
						},
						{
							Value: "3.0",
							Count: 3,
						},
					},
				},
				Pagination: struct {
					CurrentPage int `json:"currentPage"`
					PagesCount  int `json:"pagesCount"`
				}{
					CurrentPage: 1,
					PagesCount:  1,
				},
				CategoryLinks:        []any{},
				AbTestGroup:          "none",
				FilterFeatureEnabled: true,
			},
		},
	} {
		s.Run(unit.filename, func() {
			s.Require().Contains(s.payloads, unit.filename)

			got := reflect.New(reflect.TypeOf(unit.want))
			s.Require().True(FindObject(s.payloads[unit.filename], got.Interface()))
			s.Equal(unit.want, got.Elem().Interface())
		})
	}
}
