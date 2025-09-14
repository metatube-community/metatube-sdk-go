package strike3

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"golang.org/x/text/language"
	dt "gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
	"github.com/metatube-community/metatube-sdk-go/provider/strike3/internal/graphql"
)

// Constants for Strike3 provider names
const (
	BlackedRawName = "BlackedRaw"
	BlackedRawURL  = "https://www.blackedraw.com"
	BlackedName    = "Blacked"
	BlackedURL     = "https://www.blacked.com"
	VixenName      = "Vixen"
	VixenURL       = "https://www.vixen.com"
	TushyName      = "Tushy"
	TushyURL       = "https://www.tushy.com"
	TushyRawName   = "TushyRaw"
	TushyRawURL    = "https://www.tushyraw.com"
	DeeperName     = "Deeper"
	DeeperURL      = "https://www.deeper.com"
	SlayedName     = "Slayed"
	SlayedURL      = "https://www.slayed.com"
	MilfyName      = "Milfy"
	MilfyURL       = "https://www.milfy.com"
	Priority       = 1000
)

var (
	_ provider.MovieProvider = (*Strike3)(nil)
	_ provider.MovieSearcher = (*Strike3)(nil)
)

type Strike3 struct {
	*scraper.Scraper
	name     string
	siteName string
	baseURL  string
	graphQL  *graphql.Client
}

func newStrike3(name, baseURL string) *Strike3 {
	apiURL := baseURL + "/graphql"
	return &Strike3{
		name:     name,
		siteName: strings.ToUpper(name),
		baseURL:  baseURL,
		graphQL:  graphql.NewClient(apiURL),
		Scraper: scraper.NewDefaultScraper(
			name, baseURL, Priority, language.English,
		),
	}
}

func (s *Strike3) Name() string {
	return s.name
}

func (s *Strike3) SetPriority(priority float64) {
	s.Scraper.SetPriority(priority)
}

func (s *Strike3) NormalizeMovieID(id string) string {
	return strings.ToLower(id)
}

func (s *Strike3) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (s *Strike3) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return s.getMovieInfoBySlug(id)
}

func (s *Strike3) GetMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	id, err := s.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	return s.getMovieInfoBySlug(id)
}

// NormalizeMovieKeyword normalizes a movie keyword by removing site name prefixes and date prefixes
func (s *Strike3) NormalizeMovieKeyword(keyword string) string {
	return number.NormalizeMovieKeyword(keyword, s.name)
}

func (s *Strike3) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	defer func() {
		s.sortMovieSearchResults(keyword, results)
	}()

	searchKeyword := s.NormalizeMovieKeyword(keyword)

	resp, err := s.graphQL.SearchVideos(searchKeyword, s.siteName)
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.SearchVideos.Edges {
		node := edge.Node

		var releaseDate dt.Date
		if node.ReleaseDate != "" {
			if parsedDate, err := time.Parse(time.RFC3339, node.ReleaseDate); err == nil {
				releaseDate = dt.Date(parsedDate)
			}
		}

		// Get thumbnail
		var thumbURL string
		if len(node.Images.Listing) > 0 {
			thumbURL = node.Images.Listing[0].Src
		}

		results = append(results, &model.MovieSearchResult{
			ID:          node.VideoId,
			Number:      node.Slug,
			Title:       node.Title,
			Provider:    s.Name(),
			Homepage:    fmt.Sprintf("%s/videos/%s", s.baseURL, node.Slug),
			ThumbURL:    thumbURL,
			CoverURL:    thumbURL,
			ReleaseDate: releaseDate,
		})
	}

	return results, nil
}

// SearchMovie2 performs an extended search including categories and models with skip parameter
// TODO: fix this
func (s *Strike3) SearchMovie2(keyword string) (results []*model.MovieSearchResult, err error) {
	defer func() {
		s.sortMovieSearchResults(keyword, results)
	}()

	searchKeyword := s.NormalizeMovieKeyword(keyword)

	// Search using the extended GraphQL client with skip parameter
	resp, err := s.graphQL.SearchResultsTour(searchKeyword, s.siteName)
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.SearchVideos.Edges {
		node := edge.Node

		// Extract title from the node
		title := node.Title

		// Extract actors from models
		var actors []string
		for _, m := range node.Models {
			actors = append(actors, m.Name)
		}

		// Get thumbnail
		var thumbURL string
		if len(node.Images.Listing) > 0 {
			thumbURL = node.Images.Listing[0].Src
		}

		// For now, we'll use the slug as ID and number
		slug := node.Slug

		results = append(results, &model.MovieSearchResult{
			ID:       slug,
			Number:   node.Slug,
			Title:    title,
			Provider: s.Name(),
			Homepage: fmt.Sprintf("%s/videos/%s", s.baseURL, slug),
			ThumbURL: thumbURL,
			CoverURL: thumbURL, // Use thumb as cover for now
			Actors:   actors,
		})
	}

	return results, nil
}

func (s *Strike3) getMovieInfoBySlug(slug string) (*model.MovieInfo, error) {
	// Get video info using the GraphQL client
	resp, err := s.graphQL.FindOneVideo(slug, s.siteName)
	if err != nil {
		return nil, err
	}

	videoData := resp.FindOneVideo

	// Parse release date
	var releaseDate dt.Date
	if videoData.ReleaseDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, videoData.ReleaseDate); err == nil {
			releaseDate = dt.Date(parsedDate)
		}
	}

	// Extract actors
	var actors []string
	for _, m := range videoData.Models {
		actors = append(actors, m.Name)
	}

	// Extract directors
	var directors []string
	for _, director := range videoData.Directors {
		directors = append(directors, director.Name)
	}
	var director string
	if len(directors) > 0 {
		director = directors[0]
	}

	// Extract genres
	var genres []string
	for _, category := range videoData.Categories {
		genres = append(genres, category.Name)
	}

	// Extract preview images
	var previewImages []string
	for _, carousel := range videoData.Carousel {
		for _, list := range carousel.Listing {
			u := list.Highdpi.Triple
			if u != "" {
				previewImages = append(previewImages, u)
			}
		}
	}

	info := &model.MovieInfo{
		ID:            slug,
		Number:        videoData.VideoId,
		Title:         videoData.Title,
		Summary:       videoData.Description,
		Provider:      s.Name(),
		Homepage:      fmt.Sprintf("%s/videos/%s", s.baseURL, slug),
		ThumbURL:      "",
		CoverURL:      "",
		Director:      director,
		Actors:        actors,
		PreviewImages: previewImages,
		Genres:        genres,
		ReleaseDate:   releaseDate,
		Maker:         s.Name(),
		Label:         s.Name(),
		Series:        "",
		Score:         0,
	}

	// Set thumbnail and cover from preview images if available
	if len(previewImages) > 0 {
		info.ThumbURL = previewImages[0]
		info.CoverURL = previewImages[0]
		if len(previewImages) > 1 {
			info.BigCoverURL = previewImages[1]
		}
	}

	return info, nil
}

func (s *Strike3) sortMovieSearchResults(keyword string, results []*model.MovieSearchResult) {
	if len(results) == 0 {
		return
	}

	// Sort results based on similarity to the keyword
	sort.SliceStable(results, func(i, j int) bool {
		return comparer.Compare(results[i].Title, keyword) > comparer.Compare(results[j].Title, keyword)
	})
}

// BlackedRaw represents the BlackedRaw site
type BlackedRaw struct {
	*Strike3
}

func NewBlackedRaw() *BlackedRaw {
	return &BlackedRaw{
		Strike3: newStrike3(
			BlackedRawName,
			BlackedRawURL,
		),
	}
}

// Blacked represents the Blacked site
type Blacked struct {
	*Strike3
}

func NewBlacked() *Blacked {
	return &Blacked{
		Strike3: newStrike3(
			BlackedName,
			BlackedURL,
		),
	}
}

// Vixen represents the Vixen site
type Vixen struct {
	*Strike3
}

func NewVixen() *Vixen {
	return &Vixen{
		Strike3: newStrike3(
			VixenName,
			VixenURL,
		),
	}
}

// Tushy represents the Tushy site
type Tushy struct {
	*Strike3
}

func NewTushy() *Tushy {
	return &Tushy{
		Strike3: newStrike3(
			TushyName,
			TushyURL,
		),
	}
}

// TushyRaw represents the TushyRaw site
type TushyRaw struct {
	*Strike3
}

func NewTushyRaw() *TushyRaw {
	return &TushyRaw{
		Strike3: newStrike3(
			TushyRawName,
			TushyRawURL,
		),
	}
}

// Deeper represents the Deeper site
type Deeper struct {
	*Strike3
}

func NewDeeper() *Deeper {
	return &Deeper{
		Strike3: newStrike3(
			DeeperName,
			DeeperURL,
		),
	}
}

// Slayed represents the Slayed site
type Slayed struct {
	*Strike3
}

func NewSlayed() *Slayed {
	return &Slayed{
		Strike3: newStrike3(
			SlayedName,
			SlayedURL,
		),
	}
}

// Milfy represents the Milfy site
type Milfy struct {
	*Strike3
}

func NewMilfy() *Milfy {
	return &Milfy{
		Strike3: newStrike3(
			MilfyName,
			MilfyURL,
		),
	}
}

func init() {
	// Register all Strike3 Network sites
	provider.Register(BlackedRawName, NewBlackedRaw)
	provider.Register(BlackedName, NewBlacked)
	provider.Register(VixenName, NewVixen)
	provider.Register(TushyName, NewTushy)
	provider.Register(TushyRawName, NewTushyRaw)
	provider.Register(DeeperName, NewDeeper)
	provider.Register(SlayedName, NewSlayed)
	provider.Register(MilfyName, NewMilfy)
}
