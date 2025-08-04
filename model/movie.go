package model

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

const (
	MovieMetadataTableName = "movie_metadata"
	MovieReviewsTableName  = "movie_reviews"
)

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

func (m *MovieSearchResult) IsValid() bool {
	return m.ID != "" &&
		m.Number != "" &&
		m.Title != "" &&
		m.Provider != "" &&
		m.Homepage != ""
}

type MovieReviewInfo struct {
	ID          string                                   `json:"id" gorm:"primaryKey"`
	Provider    string                                   `json:"provider" gorm:"primaryKey"`
	Reviews     datatypes.JSONType[[]*MovieReviewDetail] `json:"reviews"`
	TimeTracker `json:"-"`
}

func (*MovieReviewInfo) TableName() string {
	return MovieReviewsTableName
}

func (m *MovieReviewInfo) IsValid() bool {
	if !(m.ID != "" && m.Provider != "") {
		return false
	}
	// reviews can be empty.
	for _, review := range m.Reviews.Data() {
		if !review.IsValid() {
			return false
		}
	}
	return true
}

type MovieReviewDetail struct {
	Title   string         `json:"title"`
	Author  string         `json:"author"`
	Comment string         `json:"comment"`
	Score   float64        `json:"score"`
	Date    datatypes.Date `json:"date"`
}

func (m *MovieReviewDetail) IsValid() bool {
	return m.Author != "" && m.Comment != ""
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
	return MovieMetadataTableName
}

func (m *MovieInfo) IsValid() bool {
	return m.ID != "" &&
		m.Number != "" &&
		m.Title != "" &&
		m.CoverURL != "" &&
		m.Provider != "" &&
		m.Homepage != ""
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
