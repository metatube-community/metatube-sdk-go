package muramura

import (
	"regexp"

	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/vueapi"
)

var _ provider.MovieProvider = (*MuraMura)(nil)

const (
	Name     = "MURAMURA"
	Priority = 1000 - 1
)

const (
	baseURL        = "https://www.muramura.tv/"
	movieURL       = "https://www.muramura.tv/movies/%s/"
	sampleVideoURL = "https://fms.muramura.tv/sample/%s/mb.m3u8"
)

type MuraMura struct {
	*vueapi.Core
}

func New() *MuraMura {
	return &MuraMura{
		Core: (&vueapi.Core{
			BaseURL:           baseURL,
			MovieURL:          movieURL,
			SampleVideoURL:    sampleVideoURL,
			DefaultName:       Name,
			DefaultPriority:   Priority,
			DefaultMaker:      "ムラムラってくる素人",
			GalleryPath:       "",
			LegacyGalleryPath: "",
		}).Init(),
	}
}

func (ppm *MuraMura) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
