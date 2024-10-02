package faleno

import (
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/dahlia/core"
)

var (
	_ provider.MovieProvider = (*FALENO)(nil)
	_ provider.MovieSearcher = (*FALENO)(nil)
)

const (
	Name     = "FALENO"
	Priority = 1000 - 5
)

const (
	baseURL   = "https://faleno.jp/top/"
	movieURL  = "https://faleno.jp/top/works/%s/"
	searchURL = "https://faleno.jp/top/?s=%s"
)

type FALENO struct {
	*core.Core
}

func New() *FALENO {
	return &FALENO{Core: (&core.Core{
		BaseURL:         baseURL,
		MovieURL:        movieURL,
		SearchURL:       searchURL,
		DefaultName:     Name,
		DefaultPriority: Priority,
	}).Init()}
}

func init() {
	provider.Register(Name, New)
}
