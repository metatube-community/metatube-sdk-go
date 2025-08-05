package graphql

import (
	"regexp"
)

type QueryOptions struct {
	IsLoggedIn bool
	IsAmateur  bool
	IsAnime    bool
	IsAv       bool
	IsCinema   bool
	IsSP       bool
}

var videoTypePattern = regexp.MustCompile(`//video\.dmm\.co\.jp/(\w+)/content/`)

func BuildQueryOptions(targetURL string) QueryOptions {
	var (
		typ  string
		opts QueryOptions
	)
	// E.g., https://video.dmm.co.jp/anime/content/
	if ss := videoTypePattern.FindStringSubmatch(targetURL); len(ss) == 2 {
		typ = ss[1]
	}
	switch typ {
	case "anime":
		opts.IsAnime = true
	case "amateur":
		opts.IsAmateur = true
	case "cinema":
		opts.IsCinema = true
	case "av", "vr":
		fallthrough
	default:
		opts.IsAv = true
	}
	return opts
}
