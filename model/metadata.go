package model

import (
	dt "github.com/javtube/javtube-sdk-go/model/datatypes"
)

type MovieSearchResult struct {
	ID          string  `json:"id"`
	Number      string  `json:"number"`
	Title       string  `json:"title"`
	Provider    string  `json:"provider"`
	Homepage    string  `json:"homepage"`
	ThumbURL    string  `json:"thumb_url"`
	CoverURL    string  `json:"cover_url"`
	Score       float64 `json:"score"`
	ReleaseDate dt.Date `json:"release_date"`
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

	Director string         `json:"director"`
	Actors   dt.StringArray `json:"actors"`

	ThumbURL        string         `json:"thumb_url"`
	CoverURL        string         `json:"cover_url"`
	PreviewVideoURL string         `json:"preview_video_url"`
	PreviewImages   dt.StringArray `json:"preview_images"`

	Maker     string         `json:"maker"`
	Publisher string         `json:"publisher"`
	Series    string         `json:"series"`
	Tags      dt.StringArray `json:"tags"`
	Score     float64        `json:"score"`

	Runtime     dt.Runtime `json:"runtime"`
	ReleaseDate dt.Date    `json:"release_date"`
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
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Provider string         `json:"provider"`
	Homepage string         `json:"homepage"`
	Images   dt.StringArray `json:"images"`
}

func (sr *ActorSearchResult) Valid() bool {
	return sr.ID != "" && sr.Name != "" && sr.Provider != "" && sr.Homepage != ""
}

type ActorInfo struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Provider     string         `json:"provider"`
	Homepage     string         `json:"homepage"`
	Nationality  string         `json:"nationality"`
	BloodType    string         `json:"blood_type"`
	CupSize      string         `json:"cup_size"`
	Measurements string         `json:"measurements"`
	Height       int            `json:"height"`
	Aliases      dt.StringArray `json:"aliases"`
	Images       dt.StringArray `json:"images"`
	Birthday     dt.Date        `json:"birthday"`
	DebutDate    dt.Date        `json:"debut_date"`
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
