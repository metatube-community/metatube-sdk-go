package pacopacomama

import (
	"regexp"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/1pondo/core"
)

var (
	_ provider.MovieProvider = (*Pacopacomama)(nil)
	_ provider.MovieReviewer = (*Pacopacomama)(nil)
)

const (
	Name     = "PACOPACOMAMA"
	Priority = 1000 - 1
)

const (
	baseURL        = "https://www.pacopacomama.com/"
	movieURL       = "https://www.pacopacomama.com/movies/%s/"
	sampleVideoURL = "https://fms.pacopacomama.com/sample/%s/mb.m3u8"
)

//	sampleURLs: {
//		preview: "/assets/sample/{MOVIE_ID}/s/{FILENAME}",
//		fullsize: "/assets/sample/{MOVIE_ID}/l/{FILENAME}",
//		movieIdKey: "MovieID"
//	},
const (
	galleryPath       = "/dyn/dla/images/%s"
	legacyGalleryPath = "/assets/sample/%s/l/%s"
)

type Pacopacomama struct {
	*core.Core
}

func New() *Pacopacomama {
	return &Pacopacomama{
		Core: (&core.Core{
			BaseURL:           baseURL,
			MovieURL:          movieURL,
			SampleVideoURL:    sampleVideoURL,
			DefaultName:       Name,
			DefaultPriority:   Priority,
			DefaultMaker:      "パコパコママ",
			GalleryPath:       galleryPath,
			LegacyGalleryPath: legacyGalleryPath,
		}).Init(),
	}
}

func (ppm *Pacopacomama) NormalizeMovieID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
