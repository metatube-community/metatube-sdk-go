package model

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type MovieSearchResult struct {
	ID          string         `json:"id"`
	Number      string         `json:"number"`
	Title       string         `json:"title"`
	Provider    string         `json:"provider"`
	Homepage    string         `json:"homepage"`
	ThumbURL    string         `json:"thumb_url"`
	CoverURL    string         `json:"cover_url"`
	Score       float64        `json:"score"`
	ReleaseDate datatypes.Date `json:"release_date"`
}

func (sr *MovieSearchResult) Valid() bool {
	return sr.ID != "" && sr.Number != "" && sr.Title != "" &&
		sr.Provider != "" && sr.Homepage != ""
}

type MovieInfo struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Number   string `json:"number"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Provider string `json:"provider" gorm:"primaryKey"`
	Homepage string `json:"homepage"`

	Director string         `json:"director"`
	Actors   pq.StringArray `json:"actors" gorm:"type:text[]"`

	ThumbURL        string         `json:"thumb_url"`
	CoverURL        string         `json:"cover_url"`
	PreviewVideoURL string         `json:"preview_video_url"`
	PreviewImages   pq.StringArray `json:"preview_images" gorm:"type:text[]"`

	Maker     string         `json:"maker"`
	Publisher string         `json:"publisher"`
	Series    string         `json:"series"`
	Tags      pq.StringArray `json:"tags" gorm:"type:text[]"`
	Score     float64        `json:"score"`

	Runtime     int            `json:"runtime"`
	ReleaseDate datatypes.Date `json:"release_date"`
}

func (MovieInfo) TableName() string {
	return "movie_metadata"
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
	Images   pq.StringArray `json:"images"`
}

func (sr *ActorSearchResult) Valid() bool {
	return sr.ID != "" && sr.Name != "" &&
		sr.Provider != "" && sr.Homepage != ""
}

type ActorInfo struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name"`
	Provider     string         `json:"provider" gorm:"primaryKey"`
	Homepage     string         `json:"homepage"`
	Summary      string         `json:"summary"`
	Nationality  string         `json:"nationality"`
	Hobby        string         `json:"hobby"`
	BloodType    string         `json:"blood_type"`
	CupSize      string         `json:"cup_size"`
	Measurements string         `json:"measurements"`
	Height       int            `json:"height"`
	Aliases      pq.StringArray `json:"aliases" gorm:"type:text[]"`
	Images       pq.StringArray `json:"images" gorm:"type:text[]"`
	Birthday     datatypes.Date `json:"birthday"`
	DebutDate    datatypes.Date `json:"debut_date"`
}

func (ActorInfo) TableName() string {
	return "actor_metadata"
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
