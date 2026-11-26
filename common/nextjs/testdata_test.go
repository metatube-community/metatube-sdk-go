package nextjs

import (
	"embed"
	"encoding/base64"
	"fmt"
	"testing"

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
		keys     []string
		want     map[string]string
	}{
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			keys:     []string{"src", "poster"},
			want: map[string]string{
				"src":    "https://hls.heyzo.com/sample/3000/3986/mb.m3u8",
				"type":   "application/x-mpegURL",
				"poster": "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg",
			},
		},
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			keys:     []string{"movie_seq", "movie_id"},
			want: map[string]string{
				"movie_seq": "4001",
				"movie_id":  "3986",
				"name":      "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2",
			},
		},
		{
			// https://www.dmm.co.jp/search/=/searchstr=SSIS-419/limit=30/sort=rankprofile
			filename: "a0ef9a6847f4d22cca6abc9480b7a3ae",
			keys:     []string{"search_word", "contents"},
			want: map[string]string{
				"search_word": "SSIS-419",
				"order_by":    "rankprofile",
				"limit":       "30",
				"page":        "1",
			},
		},
	} {
		s.Run(fmt.Sprintf("%s %v", unit.filename, unit.keys), func() {
			s.Require().Contains(s.payloads, unit.filename)

			var got map[string]any
			s.Require().True(FindObject(s.payloads[unit.filename], &got, unit.keys...))
			for key, want := range unit.want {
				s.Equal(want, fmt.Sprint(got[key]), key)
			}
		})
	}
}
