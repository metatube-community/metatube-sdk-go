package model

import "time"

type SearchResult struct {
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

func (sr *SearchResult) Valid() bool {
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

	Duration    time.Duration `json:"duration"`
	ReleaseDate time.Time     `json:"release_date"`
}

func (mi *MovieInfo) Valid() bool {
	return mi.ID != "" && mi.Number != "" && mi.Title != "" &&
		mi.CoverURL != "" && mi.Provider != "" && mi.Homepage != ""
}

func (mi *MovieInfo) ToSearchResult() *SearchResult {
	return &SearchResult{
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

func (asr *ActorSearchResult) Valid() bool {
	return asr.ID != "" && asr.Name != "" && asr.Provider != "" && asr.Homepage != ""
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
