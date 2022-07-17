package model

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

const MovieTableName = "movie_metadata"

// MovieSearchResult is a subset of MovieInfo.
type MovieSearchResult struct {
	ID          string         `json:"id"`
	Number      string         `json:"number"`
	Title       string         `json:"title"`
	Provider    string         `json:"provider"`
	Homepage    string         `json:"homepage"`
	ThumbURL    string         `json:"thumb_url"`
	CoverURL    string         `json:"cover_url"`
	Score       float64        `json:"score"`
	Actors      pq.StringArray `json:"actors,omitempty"`
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

	ThumbURL           string         `json:"thumb_url"`
	BigThumbURL        string         `json:"big_thumb_url"`
	CoverURL           string         `json:"cover_url"`
	BigCoverURL        string         `json:"big_cover_url"`
	PreviewVideoURL    string         `json:"preview_video_url"`
	PreviewVideoHLSURL string         `json:"preview_video_hls_url"`
	PreviewImages      pq.StringArray `json:"preview_images" gorm:"type:text[]"`

	Maker  string         `json:"maker"`
	Label  string         `json:"label"`
	Series string         `json:"series"`
	Genres pq.StringArray `json:"genres" gorm:"type:text[]"`
	Score  float64        `json:"score"`

	Runtime     int            `json:"runtime"`
	ReleaseDate datatypes.Date `json:"release_date"`

	TimeTracker `json:"-"`
}

func (*MovieInfo) TableName() string {
	return MovieTableName
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
		Actors:      m.Actors,
		ReleaseDate: m.ReleaseDate,
	}
}
