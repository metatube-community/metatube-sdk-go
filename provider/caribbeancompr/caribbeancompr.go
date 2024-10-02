package caribbeancompr

import (
	"regexp"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/caribbeancom/core"
)

var (
	_ provider.MovieProvider = (*CaribbeancomPremium)(nil)
	_ provider.MovieReviewer = (*CaribbeancomPremium)(nil)
)

const (
	Name     = "CaribbeancomPR"
	Priority = 1000 - 1
)

const (
	baseURL  = "https://www.caribbeancompr.com/"
	movieURL = "https://www.caribbeancompr.com/moviepages/%s/index.html"
)

type CaribbeancomPremium struct {
	*core.Core
}

func New() *CaribbeancomPremium {
	return &CaribbeancomPremium{
		Core: (&core.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
			DefaultMaker:    "カリビアンコムプレミアム",
		}).Init(),
	}
}

func (carib *CaribbeancomPremium) NormalizeMovieID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
