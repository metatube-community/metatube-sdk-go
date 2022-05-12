package model

import (
	"time"

	dt "github.com/javtube/javtube-sdk-go/model/datatypes"
)

type MovieSearchResult struct {
	ID          string    `json:"id"`
	Number      string    `json:"number"`
	Title       string    `json:"title"`
	Provider    string    `json:"provider"`
	Homepage    string    `json:"homepage"`
	ThumbURL    string    `json:"thumb_url"`
	CoverURL    string    `json:"cover_url"`
	Score       float64   `json:"score"`
	ReleaseDate time.Time `json:"release_date"`
}

func (sr *MovieSearchResult) Valid() bool {
	return sr.ID != "" && sr.Number != "" && sr.Title != "" && sr.Provider != "" && sr.Homepage != ""
}

type MovieInfo struct {
	ID       string `json:"id"`
	Number   string `json:"number"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Provider string `json:"provider"`
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

	Runtime     dt.Runtime `json:"runtime"`
	ReleaseDate time.Time  `json:"release_date"`
}

func (mi *MovieInfo) Valid() bool {
	return mi.ID != "" && mi.Number != "" && mi.Title != "" &&
		mi.CoverURL != "" && mi.Provider != "" && mi.Homepage != ""
}

func (mi *MovieInfo) ToSearchResult() *MovieSearchResult {
	return &MovieSearchResult{
		ID:          mi.ID,
		Number:      mi.Number,
		Title:       mi.Title,
		Provider:    mi.Provider,
		Homepage:    mi.Homepage,
		ThumbURL:    mi.ThumbURL,
		CoverURL:    mi.CoverURL,
		Score:       mi.Score,
		ReleaseDate: mi.ReleaseDate,
	}
}

type ActorSearchResult struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Provider string   `json:"provider"`
	Homepage string   `json:"homepage"`
	Images   []string `json:"images"`
}

func (sr *ActorSearchResult) Valid() bool {
	return sr.ID != "" && sr.Name != "" && sr.Provider != "" && sr.Homepage != ""
}

type ActorInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Provider     string    `json:"provider"`
	Homepage     string    `json:"homepage"`
	Nationality  string    `json:"nationality"`
	BloodType    string    `json:"blood_type"`
	CupSize      string    `json:"cup_size"`
	Measurements string    `json:"measurements"`
	Height       int       `json:"height"`
	Aliases      []string  `json:"aliases"`
	Images       []string  `json:"images"`
	Birthday     time.Time `json:"birthday"`
	DebutDate    time.Time `json:"debut_date"`
}

func (ai *ActorInfo) Valid() bool {
	return ai.ID != "" && ai.Name != "" && ai.Provider != "" && ai.Homepage != ""
}

func (ai *ActorInfo) ToSearchResult() *ActorSearchResult {
	return &ActorSearchResult{
		ID:       ai.ID,
		Name:     ai.Name,
		Provider: ai.Provider,
		Homepage: ai.Homepage,
		Images:   ai.Images,
	}
}
