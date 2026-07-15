package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

const ActorMetadataTableName = "actor_metadata"

// ActorSearchResult is a subset of ActorInfo.
type ActorSearchResult struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Provider string         `json:"provider"`
	Homepage string         `json:"homepage"`
	Aliases  pq.StringArray `json:"aliases,omitempty"`
	Images   pq.StringArray `json:"images"`
}

func (a *ActorSearchResult) IsValid() bool {
	return a.ID != "" &&
		a.Name != "" &&
		a.Provider != "" &&
		a.Homepage != ""
}

type ActorInfo struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name"`
	Provider     string         `json:"provider" gorm:"primaryKey"`
	Homepage     string         `json:"homepage"`
	Summary      string         `json:"summary"`
	Hobby        string         `json:"hobby"`
	Skill        string         `json:"skill"`
	BloodType    string         `json:"blood_type"`
	CupSize      string         `json:"cup_size"`
	Measurements string         `json:"measurements"`
	Nationality  string         `json:"nationality"`
	Height       int            `json:"height"`
	Aliases      pq.StringArray `json:"aliases" gorm:"type:text[]"`
	Images       pq.StringArray `json:"images" gorm:"type:text[]"`
	Birthday     datatypes.Date `json:"birthday"`
	DebutDate    datatypes.Date `json:"debut_date"`
	TimeTracker  `json:"-"`
}

func (*ActorInfo) TableName() string {
	return ActorMetadataTableName
}

func (a *ActorInfo) IsValid() bool {
	return a.ID != "" &&
		a.Name != "" &&
		a.Provider != "" &&
		a.Homepage != ""
}

func (a *ActorInfo) ToSearchResult() *ActorSearchResult {
	return &ActorSearchResult{
		ID:       a.ID,
		Name:     a.Name,
		Provider: a.Provider,
		Homepage: a.Homepage,
		Aliases:  a.Aliases,
		Images:   a.Images,
	}
}

// PreserveFrom fills in zero-value fields of a using non-zero values from
// existing. It implements "non-empty wins" semantics: if a has a non-zero
// value for a field, a keeps it; if a has the zero value for a field, a takes
// existing's value.
//
// This is meant to be called immediately before upserting a freshly-fetched
// ActorInfo, so that a provider response with missing optional fields does
// not clobber previously-stored good data when GORM's clause.OnConflict
// {UpdateAll: true} writes every column.
//
// Required (IsValid) fields - ID, Name, Provider, Homepage - are NOT touched:
// IsValid() gates the call and guarantees a has them all populated.
// TimeTracker fields (CreatedAt, UpdatedAt) are managed by GORM and likewise
// not touched here.
func (a *ActorInfo) PreserveFrom(existing *ActorInfo) {
	if existing == nil {
		return
	}
	// Optional string fields.
	if a.Summary == "" {
		a.Summary = existing.Summary
	}
	if a.Hobby == "" {
		a.Hobby = existing.Hobby
	}
	if a.Skill == "" {
		a.Skill = existing.Skill
	}
	if a.BloodType == "" {
		a.BloodType = existing.BloodType
	}
	if a.CupSize == "" {
		a.CupSize = existing.CupSize
	}
	if a.Measurements == "" {
		a.Measurements = existing.Measurements
	}
	if a.Nationality == "" {
		a.Nationality = existing.Nationality
	}
	// Numeric field.
	if a.Height == 0 {
		a.Height = existing.Height
	}
	// Array fields (pq.StringArray).
	if len(a.Aliases) == 0 {
		a.Aliases = existing.Aliases
	}
	if len(a.Images) == 0 {
		a.Images = existing.Images
	}
	// Date fields.
	if time.Time(a.Birthday).IsZero() {
		a.Birthday = existing.Birthday
	}
	if time.Time(a.DebutDate).IsZero() {
		a.DebutDate = existing.DebutDate
	}
}
