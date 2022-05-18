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

func (m *MovieSearchResult) Valid() bool {
	return m.ID != "" && m.Number != "" && m.Title != "" &&
		m.Provider != "" && m.Homepage != ""
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
	BigThumbURL     string         `json:"big_thumb_url"`
	BigCoverURL     string         `json:"big_cover_url"`
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

func (m *MovieInfo) Valid() bool {
	return m.ID != "" && m.Number != "" && m.Title != "" &&
		m.CoverURL != "" && m.Provider != "" && m.Homepage != ""
}

func (m *MovieInfo) ToSearchResult() *MovieSearchResult {
	return &MovieSearchResult{
		ID:          m.ID,
		Number:      m.Number,
		Title:       m.Title,
		Provider:    m.Provider,
		Homepage:    m.Homepage,
		ThumbURL:    m.ThumbURL,
		CoverURL:    m.CoverURL,
		Score:       m.Score,
		ReleaseDate: m.ReleaseDate,
	}
}

type ActorSearchResult struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Provider string         `json:"provider"`
	Homepage string         `json:"homepage"`
	Images   pq.StringArray `json:"images"`
}

func (a *ActorSearchResult) Valid() bool {
	return a.ID != "" && a.Name != "" &&
		a.Provider != "" && a.Homepage != ""
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

func (a *ActorInfo) Valid() bool {
	return a.ID != "" && a.Name != "" && a.Provider != "" && a.Homepage != ""
}

func (a *ActorInfo) ToSearchResult() *ActorSearchResult {
	return &ActorSearchResult{
		ID:       a.ID,
		Name:     a.Name,
		Provider: a.Provider,
		Homepage: a.Homepage,
		Images:   a.Images,
	}
}
