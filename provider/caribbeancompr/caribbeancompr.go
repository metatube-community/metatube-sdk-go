package caribbeancompr

import (
	"regexp"

	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/caribapi"
)

var _ provider.MovieProvider = (*CaribbeancomPremium)(nil)

const (
	Name     = "CaribbeancomPR"
	Priority = 1000 - 1
)

const (
	baseURL  = "https://www.caribbeancompr.com/"
	movieURL = "https://www.caribbeancompr.com/moviepages/%s/index.html"
)

type CaribbeancomPremium struct {
	*caribapi.Core
}

func New() *CaribbeancomPremium {
	return &CaribbeancomPremium{
		Core: (&caribapi.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
			DefaultMaker:    "カリビアンコムプレミアム",
		}).Init(),
	}
}

func (carib *CaribbeancomPremium) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
