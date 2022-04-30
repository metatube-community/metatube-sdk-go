package model

import "time"

type SearchResult struct {
	ID          string    `json:"id"`
	Number      string    `json:"number"`
	Title       string    `json:"title"`
	ThumbURL    string    `json:"thumb_url"`
	CoverURL    string    `json:"cover_url"`
	ReleaseDate time.Time `json:"release_date"`
}

type MovieInfo struct {
	ID       string `json:"id"`
	Number   string `json:"number"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Homepage string `json:"homepage"`

	Director string   `json:"director"`
	Actors   []string `json:"actors"`

	ThumbURL        string   `json:"thumb_url"`
	CoverURL        string   `json:"cover_url"`
	PreviewVideoURL string   `json:"preview_video_url"`
	PreviewImages   []string `json:"preview_images"`

	Maker     string   `json:"maker"`
	Publisher string   `json:"publisher"`
	Series    string   `json:"series"`
	Tags      []string `json:"tags"`
	Score     float64  `json:"score"`

	Duration    time.Duration `json:"duration"`
	ReleaseDate time.Time     `json:"release_date"`
}
