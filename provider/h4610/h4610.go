package h4610

import (
	"regexp"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/h0930/core"
)

var _ provider.MovieProvider = (*H4610)(nil)

const (
	Name     = "H4610"
	Priority = 1000
)

const (
	baseURL  = "https://www.h4610.com/"
	movieURL = "https://www.h4610.com/moviepages/%s/index.html"
)

type H4610 struct {
	*core.Core
}

func New() *H4610 {
	return &H4610{
		Core: (&core.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
			DefaultMaker:    "エッチな4610",
		}).Init(),
	}
}

func (h *H4610) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:h4610[-_])?([a-z\d]+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return strings.ToLower(ss[1])
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
