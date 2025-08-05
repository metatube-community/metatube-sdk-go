package graphql

import (
	"net/url"
	"strings"
)

type QueryOptions struct {
	IsLoggedIn bool
	IsAmateur  bool
	IsAnime    bool
	IsAv       bool
	IsCinema   bool
	IsSP       bool
}

func GenerateQueryOptions(u *url.URL) QueryOptions {
	// E.g., https://video.dmm.co.jp/anime/content/
	typ := strings.SplitN(
		strings.TrimPrefix(u.Path, "/"),
		"/", 2,
	)[0]
	opts := QueryOptions{}
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
