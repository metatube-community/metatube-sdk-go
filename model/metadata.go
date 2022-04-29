package model

import "time"

type SearchResult struct {
	ID          string
	Number      string
	Title       string
	ThumbURL    string
	CoverURL    string
	ReleaseDate time.Time
}

type MovieInfo struct {
	ID       string
	Number   string
	Title    string
	Summary  string
	Homepage string

	Director string
	Actors   []struct {
		ID   string
		Name string
	}

	ThumbURL        string
	CoverURL        string
	PreviewVideoURL string
	PreviewImages   []string

	Maker     string
	Publisher string
	Series    string
	Tags      []string
	Score     float64

	Duration    time.Duration
	ReleaseDate time.Time
}
