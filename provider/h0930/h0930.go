package h0930

import (
	"regexp"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/h0930/core"
)

var _ provider.MovieProvider = (*H0930)(nil)

const (
	Name     = "H0930"
	Priority = 1000
)

const (
	baseURL  = "https://www.h0930.com/"
	movieURL = "https://www.h0930.com/moviepages/%s/index.html"
)

type H0930 struct {
	*core.Core
}

func New() *H0930 {
	return &H0930{
		Core: (&core.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
			DefaultMaker:    "エッチな0930",
		}).Init(),
	}
}

func (h *H0930) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:h0930[-_])?([a-z\d]+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return strings.ToLower(ss[1])
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
