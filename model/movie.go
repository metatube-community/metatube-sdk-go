package model

import (
	"time"

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
	ID          string                                  `json:"id" gorm:"primaryKey"`
	Provider    string                                  `json:"provider" gorm:"primaryKey"`
	Reviews     datatypes.JSONSlice[*MovieReviewDetail] `json:"reviews"`
	TimeTracker `json:"-"`
}

func (*MovieReviewInfo) TableName() string {
	return MovieReviewsTableName
}

func (m *MovieReviewInfo) IsValid() bool {
	if m.ID == "" || m.Provider == "" {
		return false
	}
	// reviews can be empty.
	for _, review := range m.Reviews {
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

// PreserveFrom fills in zero-value fields of m using non-zero values from
// existing. It implements "non-empty wins" semantics: if m has a non-zero
// value for a field, m keeps it; if m has the zero value for a field, m takes
// existing's value.
//
// This is meant to be called immediately before upserting a freshly-fetched
// MovieInfo, so that a provider response with missing optional fields does
// not clobber previously-stored good data when GORM's clause.OnConflict
// {UpdateAll: true} writes every column.
//
// Required (IsValid) fields - ID, Number, Title, CoverURL, Provider, Homepage -
// are NOT touched: IsValid() gates the call and guarantees m has them all
// populated, and the primary key columns (ID, Provider) must come from m anyway.
// TimeTracker fields (CreatedAt, UpdatedAt) are NOT touched either: GORM
// manages them automatically via AutoCreateTime / AutoUpdateTime, and they
// are special-cased out of the OnConflict update.
func (m *MovieInfo) PreserveFrom(existing *MovieInfo) {
	if existing == nil {
		return
	}
	// Optional string fields.
	if m.Summary == "" {
		m.Summary = existing.Summary
	}
	if m.Director == "" {
		m.Director = existing.Director
	}
	if m.ThumbURL == "" {
		m.ThumbURL = existing.ThumbURL
	}
	if m.BigThumbURL == "" {
		m.BigThumbURL = existing.BigThumbURL
	}
	if m.BigCoverURL == "" {
		m.BigCoverURL = existing.BigCoverURL
	}
	if m.PreviewVideoURL == "" {
		m.PreviewVideoURL = existing.PreviewVideoURL
	}
	if m.PreviewVideoHLSURL == "" {
		m.PreviewVideoHLSURL = existing.PreviewVideoHLSURL
	}
	if m.Maker == "" {
		m.Maker = existing.Maker
	}
	if m.Label == "" {
		m.Label = existing.Label
	}
	if m.Series == "" {
		m.Series = existing.Series
	}
	// Array fields (pq.StringArray).
	if len(m.Actors) == 0 {
		m.Actors = existing.Actors
	}
	if len(m.PreviewImages) == 0 {
		m.PreviewImages = existing.PreviewImages
	}
	if len(m.Genres) == 0 {
		m.Genres = existing.Genres
	}
	// Numeric fields.
	if m.Score == 0 {
		m.Score = existing.Score
	}
	if m.Runtime == 0 {
		m.Runtime = existing.Runtime
	}
	// Date field.
	if time.Time(m.ReleaseDate).IsZero() {
		m.ReleaseDate = existing.ReleaseDate
	}
}
