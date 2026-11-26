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
				Src               string `json:"src"`
				Type              string `json:"type"`
				Poster            string `json:"poster"`
				SeekPreviewVTTURL string `json:"seekPreviewVttUrl"`
				MovieSeq          int    `json:"movieSeq"`
				IsHeyzoMember     bool   `json:"isHeyzoMember"`
			}{
				Src:               "https://hls.heyzo.com/sample/3000/3986/mb.m3u8",
				Type:              "application/x-mpegURL",
				Poster:            "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg",
				SeekPreviewVTTURL: "",
				MovieSeq:          4001,
				IsHeyzoMember:     false,
			},
		},
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			want: struct {
				MovieSeq         int       `json:"movie_seq"`
				MovieID          string    `json:"movie_id"`
				Name             string    `json:"name"`
				MonthlyStartDate time.Time `json:"monthly_start_date"`
				Detail           struct {
					RvScore float64 `json:"rv_score"`
					GroupID int     `json:"group_id"`
				} `json:"detail"`
				Actors   []string `json:"actors"`
				Download struct {
					Whole []struct {
						Duration   string `json:"duration"`
						FileSizeMb int    `json:"file_size_mb"`
					} `json:"whole"`
					Split []any `json:"split"`
				} `json:"download"`
				Count struct {
					Reviews int `json:"reviews"`
				} `json:"_count"`
			}{
				MovieSeq:         4001,
				MovieID:          "3986",
				Name:             "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2",
				MonthlyStartDate: time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC),
				Detail: struct {
					RvScore float64 `json:"rv_score"`
					GroupID int     `json:"group_id"`
				}{
					RvScore: 0,
					GroupID: 2,
				},
				Actors: []string{
					"$6:props:children:1:props:children:0:props:actor",
				},
				Download: struct {
					Whole []struct {
						Duration   string `json:"duration"`
						FileSizeMb int    `json:"file_size_mb"`
					} `json:"whole"`
					Split []any `json:"split"`
				}{
					Whole: []struct {
						Duration   string `json:"duration"`
						FileSizeMb int    `json:"file_size_mb"`
					}{
						{
							Duration:   "00:59:48",
							FileSizeMb: 216,
						},
					},
					Split: []any{},
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
				OrderBy    string `json:"order_by"`
				Limit      int    `json:"limit"`
				Page       int    `json:"page"`
				Pagination struct {
					CurrentPage int `json:"currentPage"`
					PagesCount  int `json:"pagesCount"`
				} `json:"pagination"`
				CategoryLinks []any `json:"category_links"`
				Contents      struct {
					Count int `json:"count"`
					Data  []struct {
						ContentID string   `json:"content_id"`
						Casts     []string `json:"casts"`
						Rate      float64  `json:"rate"`
						Authors   []string `json:"authors"`
					} `json:"data"`
				} `json:"contents"`
			}{
				SearchWord: "SSIS-419",
				OrderBy:    "rankprofile",
				Limit:      30,
				Page:       1,
				Pagination: struct {
					CurrentPage int `json:"currentPage"`
					PagesCount  int `json:"pagesCount"`
				}{
					CurrentPage: 1,
					PagesCount:  1,
				},
				CategoryLinks: []any{},
				Contents: struct {
					Count int `json:"count"`
					Data  []struct {
						ContentID string   `json:"content_id"`
						Casts     []string `json:"casts"`
						Rate      float64  `json:"rate"`
						Authors   []string `json:"authors"`
					} `json:"data"`
				}{
					Count: 3,
					Data: []struct {
						ContentID string   `json:"content_id"`
						Casts     []string `json:"casts"`
						Rate      float64  `json:"rate"`
						Authors   []string `json:"authors"`
					}{
						{
							ContentID: "ssis419",
							Casts: []string{
								"三上悠亜",
							},
							Rate:    3,
							Authors: nil,
						},
						{
							ContentID: "9ssis419",
							Casts: []string{
								"三上悠亜",
							},
							Rate:    4,
							Authors: nil,
						},
						{
							ContentID: "4ssis419",
							Casts: []string{
								"三上悠亜",
							},
							Rate:    3,
							Authors: nil,
						},
					},
				},
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
