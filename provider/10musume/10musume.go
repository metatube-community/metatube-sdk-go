package tenmusume

import (
	"regexp"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/1pondo/core"
)

var (
	_ provider.MovieProvider = (*TenMusume)(nil)
	_ provider.MovieReviewer = (*TenMusume)(nil)
)

const (
	Name     = "10musume"
	Priority = 1000
)

const (
	baseURL        = "https://www.10musume.com/"
	movieURL       = "https://www.10musume.com/movies/%s/"
	sampleVideoURL = "https://fms.10musume.com/sample/%s/mb.m3u8"
)

//	sampleURLs: {
//	  preview: "/assets/sample/{MOVIE_ID}/{FILENAME}",
//	  fullsize: "/assets/sample/{MOVIE_ID}/{FILENAME}",
//	  movieIdKey: "MovieID"
//	}
const (
	galleryPath       = "/dyn/dla/images/%s"
	legacyGalleryPath = "/assets/sample/%s/%s"
)

type TenMusume struct {
	*core.Core
}

func New() *TenMusume {
	return &TenMusume{
		Core: (&core.Core{
			BaseURL:           baseURL,
			MovieURL:          movieURL,
			SampleVideoURL:    sampleVideoURL,
			DefaultName:       Name,
			DefaultPriority:   Priority,
			DefaultMaker:      "天然むすめ",
			GalleryPath:       galleryPath,
			LegacyGalleryPath: legacyGalleryPath,
		}).Init(),
	}
}

func (mse *TenMusume) NormalizeMovieID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{2}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
