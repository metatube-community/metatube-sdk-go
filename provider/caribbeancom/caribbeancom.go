package caribbeancom

import (
	"regexp"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/caribbeancom/core"
)

var (
	_ provider.MovieProvider = (*Caribbeancom)(nil)
	_ provider.MovieReviewer = (*Caribbeancom)(nil)
)

const (
	Name     = "Caribbeancom"
	Priority = 1000
)

const (
	baseURL  = "https://www.caribbeancom.com/"
	movieURL = "https://www.caribbeancom.com/moviepages/%s/index.html"
)

type Caribbeancom struct {
	*core.Core
}

func New() *Caribbeancom {
	return &Caribbeancom{
		Core: (&core.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
			DefaultMaker:    "カリビアンコム",
		}).Init(),
	}
}

func (carib *Caribbeancom) NormalizeMovieID(id string) string {
	if regexp.MustCompile(`^\d{6}-\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
