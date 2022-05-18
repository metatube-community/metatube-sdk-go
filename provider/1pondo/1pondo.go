package onepondo

import (
	"regexp"

	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/d2pass"
)

var _ provider.MovieProvider = (*OnePondo)(nil)

const (
	Name     = "1PONDO"
	Priority = 1000
)

const (
	baseURL  = "https://www.1pondo.tv/"
	movieURL = "https://www.1pondo.tv/movies/%s/"
)

//sampleURLs: {
//   preview: "/assets/sample/{MOVIE_ID}/thum_106/{FILENAME}.jpg",
//   fullsize: "/assets/sample/{MOVIE_ID}/popu/{FILENAME}.jpg",
//   movieIdKey: "MovieID"
//}
const (
	galleryPath       = "/dyn/dla/images/%s"
	legacyGalleryPath = "/assets/sample/%s/popu/%s"
)

type OnePondo struct {
	*d2pass.Core
}

func New() *OnePondo {
	return &OnePondo{
		Core: (&d2pass.Core{
			BaseURL:           baseURL,
			MovieURL:          movieURL,
			DefaultName:       Name,
			DefaultPriority:   Priority,
			DefaultMaker:      "一本道",
			GalleryPath:       galleryPath,
			LegacyGalleryPath: legacyGalleryPath,
		}).Init(),
	}
}

func (opd *OnePondo) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
