package onepondo

import (
	"regexp"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/1pondo/core"
)

var (
	_ provider.MovieProvider = (*OnePondo)(nil)
	_ provider.MovieReviewer = (*OnePondo)(nil)
	_ provider.Fetcher       = (*OnePondo)(nil)
)

const (
	Name     = "1Pondo"
	Priority = 1000
)

const (
	baseURL        = "https://www.1pondo.tv/"
	movieURL       = "https://www.1pondo.tv/movies/%s/"
	sampleVideoURL = "https://fms.1pondo.tv/sample/%s/mb.m3u8"
)

//	sampleURLs: {
//	  preview: "/assets/sample/{MOVIE_ID}/thum_106/{FILENAME}.jpg",
//	  fullsize: "/assets/sample/{MOVIE_ID}/popu/{FILENAME}.jpg",
//	  movieIdKey: "MovieID"
//	}
const (
	galleryPath       = "/dyn/dla/images/%s"
	legacyGalleryPath = "/assets/sample/%s/popu/%s"
)

type OnePondo struct {
	*core.Core
}

func New() *OnePondo {
	return &OnePondo{
		Core: (&core.Core{
			BaseURL:           baseURL,
			MovieURL:          movieURL,
			SampleVideoURL:    sampleVideoURL,
			DefaultName:       Name,
			DefaultPriority:   Priority,
			DefaultMaker:      "一本道",
			GalleryPath:       galleryPath,
			LegacyGalleryPath: legacyGalleryPath,
		}).Init(),
	}
}

func (opd *OnePondo) NormalizeMovieID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
