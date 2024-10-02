package dahlia

import (
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/dahlia/core"
)

var (
	_ provider.MovieProvider = (*DAHLIA)(nil)
	_ provider.MovieSearcher = (*DAHLIA)(nil)
)

const (
	Name     = "DAHLIA"
	Priority = 1000 - 5
)

const (
	baseURL   = "https://dahlia-av.jp/"
	movieURL  = "https://dahlia-av.jp/works/%s/"
	searchURL = "https://dahlia-av.jp/?s=%s"
)

type DAHLIA struct {
	*core.Core
}

func New() *DAHLIA {
	return &DAHLIA{
		Core: (&core.Core{
			BaseURL:         baseURL,
			MovieURL:        movieURL,
			SearchURL:       searchURL,
			DefaultName:     Name,
			DefaultPriority: Priority,
		}).Init(),
	}
}

func init() {
	provider.Register(Name, New)
}
