package nextjs

import (
	"embed"
	"encoding/base64"
	"reflect"
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
		want     any // struct value to find, matched by its fields.
	}{
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			want: struct {
				Src    string `json:"src"`
				Type   string `json:"type"`
				Poster string `json:"poster"`
			}{
				Src:    "https://hls.heyzo.com/sample/3000/3986/mb.m3u8",
				Type:   "application/x-mpegURL",
				Poster: "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg",
			},
		},
		{
			// https://www.heyzo.com/movie/3986
			filename: "55307bfd703df11319fecbfed69883f5",
			want: struct {
				MovieSeq int    `json:"movie_seq"`
				MovieID  string `json:"movie_id"`
				Name     string `json:"name"`
			}{
				MovieSeq: 4001,
				MovieID:  "3986",
				Name:     "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2",
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
			}{
				SearchWord: "SSIS-419",
				OrderBy:    "rankprofile",
				Limit:      30,
				Page:       1,
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
