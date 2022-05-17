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

const baseURL = "https://www.1pondo.tv/"

const (
	galleryPath       = "/dyn/dla/images/%s"
	legacyGalleryPath = "/assets/sample/%s/popu/%s"
)

type OnePondo struct {
	*d2pass.Core
}

func New() *OnePondo {
	core := &d2pass.Core{
		BaseURL:           baseURL,
		DefaultName:       Name,
		DefaultPriority:   Priority,
		DefaultMaker:      "一本道",
		GalleryPath:       galleryPath,
		LegacyGalleryPath: legacyGalleryPath,
	}
	core.Init()
	return &OnePondo{core}
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
