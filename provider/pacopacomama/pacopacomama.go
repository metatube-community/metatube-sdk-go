package pacopacomama

import (
	"regexp"

	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/vue"
)

var _ provider.MovieProvider = (*Pacopacomama)(nil)

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
	*vue.Core
}

func New() *Pacopacomama {
	return &Pacopacomama{
		Core: (&vue.Core{
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

func (ppm *Pacopacomama) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
