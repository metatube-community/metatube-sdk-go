package model

import (
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
