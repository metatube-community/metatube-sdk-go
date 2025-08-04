package theporndb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*ThePornDBVideo)(nil)
	_ provider.MovieSearcher = (*ThePornDBVideo)(nil)
)

const (
	SceneProviderName = "ThePornDBScene"
	MovieProviderName = "ThePornDBMovie"

	movieBaseURL = "https://theporndb.net/movies/"
	sceneBaseURL = "https://theporndb.net/scenes/"

	moviePageURL = "https://theporndb.net/movies/%s"
	scenePageURL = "https://theporndb.net/scenes/%s"

	apiGetMovieURL  = "https://api.theporndb.net/movies/%s"
	apiGetScenesURL = "https://api.theporndb.net/scenes/%s"

	apiSearchMovieURL = "https://api.theporndb.net/movies?q=%s"
	apiSearchSceneURL = "https://api.theporndb.net/scenes?q=%s"
)

// ThePornDB have different API endpoints for scenes and movies, but response
// JSON schema is the same. So we need to register 2 video providers.

type ThePornDBVideo struct {
	*scraper.Scraper

	pageURL      string
	apiGetURL    string
	apiSearchURL string

	accessToken string
}

func new(name, baseURL, pageURL, apiGetURL, apiSearchURL string) *ThePornDBVideo {
	return &ThePornDBVideo{
		Scraper:      scraper.NewDefaultScraper(name, baseURL, Priority, language.English),
		pageURL:      pageURL,
		apiGetURL:    apiGetURL,
		apiSearchURL: apiSearchURL,
		accessToken:  "",
	}
}

func NewThePornDBScene() *ThePornDBVideo {
	return new(SceneProviderName, sceneBaseURL, scenePageURL, apiGetScenesURL, apiSearchSceneURL)
}

func NewThePornDBMovie() *ThePornDBVideo {
	return new(MovieProviderName, movieBaseURL, moviePageURL, apiGetMovieURL, apiSearchMovieURL)
}

func (s *ThePornDBVideo) SetConfig(config map[string]string) error {
	if accessToken, ok := config["ACCESS_TOKEN"]; ok {
		fmt.Println(s.Name(), "set token")
		s.accessToken = accessToken
	}
	return nil
}

// GetMovieInfoByID impls MovieProvider.GetMovieInfoByID.
func (s *ThePornDBVideo) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	if s.accessToken == "" {
		return nil, nil
	}

	info = &model.MovieInfo{
		Provider:      s.Name(),
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := s.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := &getVideoResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}

		info.ID = resp.Data.Slug
		info.Number = resp.Data.Slug
		info.Homepage = fmt.Sprintf(s.pageURL, info.ID)
		info.Title = resp.Data.Title
		info.Summary = resp.Data.Description
		info.ThumbURL = resp.Data.Poster
		info.CoverURL = resp.Data.Image
		info.Score = float64(resp.Data.Rating)
		info.PreviewVideoURL = resp.Data.Trailer
		info.Maker = resp.Data.Site.Name
		info.Runtime = resp.Data.Duration

		if releaseDate, err := resp.Data.ReleaseDate(); err == nil {
			info.ReleaseDate = releaseDate
		}

		for _, tag := range resp.Data.Tags {
			info.Genres = append(info.Genres, tag.Name)
		}

		for _, actor := range resp.Data.Performers {
			info.Actors = append(info.Actors, actor.Name)
		}

		if len(resp.Data.Directors) > 0 {
			info.Director = resp.Data.Directors[0].Name
		}
	})

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	err = c.Request(http.MethodGet, fmt.Sprintf(s.apiGetURL, id), nil, nil, headers)
	return
}

// ParseMovieIDFromURL impls MovieProvider.ParseMovieIDFromURL.
func (s *ThePornDBVideo) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

// GetMovieInfoByURL impls MovieProvider.GetMovieInfoByURL.
func (s *ThePornDBVideo) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := s.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	return s.GetMovieInfoByID(id)
}

// NormalizeMovieKeyword impls MovieSearcher.NormalizeMovieKeyword.
func (s *ThePornDBVideo) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

// SearchMovie impls MovieSearcher.SearchMovie.
func (s *ThePornDBVideo) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	if s.accessToken == "" {
		return nil, nil
	}

	c := s.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := &searchVideosResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}
		for _, video := range resp.Data {
			releaseDate, _ := video.ReleaseDate()
			results = append(results, &model.MovieSearchResult{
				ID:          video.Slug,
				Number:      video.Slug,
				Title:       video.Title,
				Provider:    s.Name(),
				Homepage:    fmt.Sprintf(s.apiGetURL, video.Slug),
				ThumbURL:    video.Poster,
				CoverURL:    video.Image,
				ReleaseDate: releaseDate,
			})
		}
	})

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	err = c.Request(http.MethodGet, fmt.Sprintf(s.apiSearchURL, url.QueryEscape(keyword)), nil, nil, headers)
	return
}
