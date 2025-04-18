package theporndb

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/datatypes"
)

type actorInfo struct {
	ID        string  `json:"id"`
	NumericID uint64  `json:"_id"`
	Slug      string  `json:"slug"`
	Name      string  `json:"name"`
	Bio       string  `json:"bio"`
	Rating    float32 `json:"rating"`
	IsParent  bool    `json:"is_parent"`
	Extras    struct {
		Gender            string `json:"gender"`
		Birthday          string `json:"birthday"`
		BirthdayTimestamp int    `json:"birthday_timestamp"`
		Birthplace        string `json:"birthplace"`
		BirthplaceCode    string `json:"birthplace_code"`
		Astrology         string `json:"astrology"`
		Ethnicity         string `json:"ethnicity"`
		Nationality       string `json:"nationality"`
		HairColour        string `json:"hair_colour"`
		EyeColour         string `json:"eye_colour"`
		Weight            string `json:"weight"`
		Height            string `json:"height"`
		Measurements      string `json:"measurements"`
		Cupsize           string `json:"cupsize"`
		Tattoos           string `json:"tattoos"`
		Piercings         string `json:"piercings"`
		Waist             string `json:"waist"`
		Hips              string `json:"hips"`
		FakeBoobs         bool   `json:"fake_boobs"`
		SameSexOnly       bool   `json:"same_sex_only"`
		CareerStartYear   int    `json:"career_start_year"`
		CareerEndYear     int    `json:"career_end_year"`
	} `json:"extras"`
	Aliases   []string `json:"aliases"`
	Image     string   `json:"image"`
	Thumbnail string   `json:"thumbnail"`
	Face      string   `json:"face"`
	Posters   []struct {
		ID    int    `json:"id"`
		URL   string `json:"url"`
		Size  int    `json:"size"`
		Order int    `json:"order"`
	} `json:"posters"`
}

func (a *actorInfo) BirthdayDate() (datatypes.Date, error) {
	parsedDate, err := time.Parse("2006-01-02", a.Extras.Birthday)
	return datatypes.Date(parsedDate), err
}

func (a *actorInfo) HeightInCM() int {
	if a.Extras.Height == "" {
		return 0
	}
	if !strings.HasSuffix(a.Extras.Height, "cm") {
		return 0
	}
	h, _ := strconv.Atoi(strings.TrimSuffix(a.Extras.Height, "cm"))
	return h
}

type getActorResponse struct {
	Data actorInfo `json:"data"`
}

type searchActorResponse struct {
	Data  []actorInfo `json:"data"`
	Links struct {
		First string      `json:"first"`
		Last  string      `json:"last"`
		Prev  interface{} `json:"prev"`
		Next  interface{} `json:"next"`
	} `json:"links"`
	Meta struct {
		CurrentPage int `json:"current_page"`
		From        int `json:"from"`
		LastPage    int `json:"last_page"`
		Links       []struct {
			URL    interface{} `json:"url"`
			Label  string      `json:"label"`
			Active bool        `json:"active"`
		} `json:"links"`
		Path    string `json:"path"`
		PerPage int    `json:"per_page"`
		To      int    `json:"to"`
		Total   int    `json:"total"`
	} `json:"meta"`
}

type videoInfo struct {
	ID         string `json:"id"`
	NumericID  uint64 `json:"_id"`
	ExternalID string `json:"external_id"`
	// Slug is the meaningful id.
	Slug string `json:"slug"`

	Title       string  `json:"title"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Rating      float32 `json:"rating"`
	SiteID      int     `json:"site_id"`
	Date        string  `json:"date"`
	URL         string  `json:"url"`

	// Cover
	Image       string `json:"image"`
	BackImage   string `json:"back_image"`
	PosterImage string `json:"poster_image"`
	// Thumbnail
	Poster   string `json:"poster"`
	Trailer  string `json:"trailer"`
	Duration int    `json:"duration"`

	Performers []actorInfo `json:"performers"`
	Site       struct {
		Name string `json:"name"`
	} `json:"site"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
	Directors []struct {
		Name string `json:"name"`
	} `json:"directors"`
}

func (v *videoInfo) ReleaseDate() (datatypes.Date, error) {
	parsedDate, err := time.Parse("2006-01-02", v.Date)
	if err != nil {
		return datatypes.Date{}, err
	}
	return datatypes.Date(parsedDate), nil
}

type getVideoResponse struct {
	Data videoInfo `json:"data"`
}

type searchVideosResponse struct {
	Data  []videoInfo `json:"data"`
	Links struct {
		First string      `json:"first"`
		Last  string      `json:"last"`
		Prev  interface{} `json:"prev"`
		Next  interface{} `json:"next"`
	} `json:"links"`
	Meta struct {
		CurrentPage int `json:"current_page"`
		From        int `json:"from"`
		LastPage    int `json:"last_page"`
		Links       []struct {
			URL    interface{} `json:"url"`
			Label  string      `json:"label"`
			Active bool        `json:"active"`
		} `json:"links"`
		Path    string `json:"path"`
		PerPage int    `json:"per_page"`
		To      int    `json:"to"`
		Total   int    `json:"total"`
	} `json:"meta"`
}
