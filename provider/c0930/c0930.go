package c0930

import (
	"regexp"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/h0930/core"
)

var _ provider.MovieProvider = (*C0930)(nil)

const (
	Name     = "C0930"
	Priority = 1000
)

const (
	baseURL  = "https://www.c0930.com/"
	movieURL = "https://www.c0930.com/moviepages/%s/index.html"
)

type C0930 struct {
	*core.Core
}

func New() *C0930 {
	return &C0930{
		Core: (&core.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
			DefaultMaker:    "人妻斬り",
		}).Init(),
	}
}

func (h *C0930) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:c0930[-_])?([a-z\d]+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return strings.ToLower(ss[1])
	}
	return ""
}

func init() {
	provider.Register(Name, New)
}
