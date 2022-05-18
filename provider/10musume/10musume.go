package tenmusume

import (
	"regexp"

	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/d2pass"
)

var _ provider.MovieProvider = (*TenMusume)(nil)

const (
	Name     = "10MUSUME"
	Priority = 1000
)

const (
	baseURL  = "https://www.10musume.com/"
	movieURL = "https://www.10musume.com/movies/%s/"
)

// sampleURLs: {
//   preview: "/assets/sample/{MOVIE_ID}/{FILENAME}",
//   fullsize: "/assets/sample/{MOVIE_ID}/{FILENAME}",
//   movieIdKey: "MovieID"
//}
const (
	galleryPath       = "/dyn/dla/images/%s"
	legacyGalleryPath = "/assets/sample/%s/%s"
)

type TenMusume struct {
	*d2pass.Core
}

func New() *TenMusume {
	return &TenMusume{
		Core: (&d2pass.Core{
			BaseURL:           baseURL,
			MovieURL:          movieURL,
			DefaultName:       Name,
			DefaultPriority:   Priority,
			DefaultMaker:      "天然むすめ",
			GalleryPath:       galleryPath,
			LegacyGalleryPath: legacyGalleryPath,
		}).Init(),
	}
}

func (mse *TenMusume) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{2}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
