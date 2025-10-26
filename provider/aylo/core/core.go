package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*Aylo)(nil)
	_ provider.MovieSearcher = (*Aylo)(nil)
)

const (
	APIBaseURL = "https://site-api.project1service.com"
	APIVersion = "v2"

	ReleaseEndpoint = "/releases/%s"
	SearchEndpoint  = "/releases?search=%s&type=scene&limit=20"
	priority        = 1000
	userAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:79.0) Gecko/20100101 Firefox/79.0"
)

type Aylo struct {
	*scraper.Scraper
	baseURL     string
	token       string
	tokenExpire time.Time
}

func New(name, brand string) *Aylo {
	baseURL := fmt.Sprintf("https://www.%s.com", brand)
	return &Aylo{
		Scraper: scraper.NewDefaultScraper(name, baseURL, priority, language.English),
		baseURL: baseURL,
		token:   "",
	}
}

// GetMovieInfoByID implements MovieProvider.GetMovieInfoByID
func (a *Aylo) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	// Ensure token is valid before making request
	if err := a.getTokenWithAutoRefresh(); err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		Provider:      a.Name(),
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	apiURL := fmt.Sprintf("%s/%s%s", APIBaseURL, APIVersion, fmt.Sprintf(ReleaseEndpoint, id))

	// Create request with headers
	headers := http.Header{}
	headers.Set("Instance", a.token)

	headers.Set("User-Agent", userAgent)
	headers.Set("Origin", a.baseURL)
	headers.Set("Referer", a.baseURL)

	// Create a new collector for this request
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		var apiResp struct {
			Result map[string]interface{} `json:"result"`
		}

		if err = json.Unmarshal(r.Body, &apiResp); err != nil {
			return
		}

		result := apiResp.Result

		// If this is a trailer, get the parent scene data
		if result["type"] != "scene" {
			if parent, ok := result["parent"].(map[string]interface{}); ok && parent["type"] == "scene" {
				result = parent
			}
		}

		// Parse the scene data
		info.ID = fmt.Sprintf("%v", result["id"])
		info.Number = info.ID
		info.Title = fmt.Sprintf("%v", result["title"])

		if desc, ok := result["description"].(string); ok {
			info.Summary = desc
		}

		// Handle date
		if dateStr, ok := result["dateReleased"].(string); ok {
			if date, err := time.Parse(time.RFC3339, dateStr); err == nil {
				info.ReleaseDate = datatypes.Date(date)
			}
		}

		// Handle images
		if images, ok := result["images"].(map[string]interface{}); ok {
			if poster, ok := images["poster"].(map[string]interface{}); ok {
				if first, ok := poster["0"].(map[string]interface{}); ok {
					if xx, ok := first["xx"].(map[string]interface{}); ok {
						if url, ok := xx["url"].(string); ok {
							info.ThumbURL = strings.Split(url, "/m=")[0] + "/m=COV_400x600"
							info.CoverURL = strings.Split(url, "/m=")[0] + "/m=COV_400x600"
						}
					}
				}
			}
		}

		// Handle actors
		if actors, ok := result["actors"].([]interface{}); ok {
			for _, actor := range actors {
				if actorMap, ok := actor.(map[string]interface{}); ok {
					if name, ok := actorMap["name"].(string); ok {
						info.Actors = append(info.Actors, name)
					}
				}
			}
		}

		// Handle tags
		if tags, ok := result["tags"].([]interface{}); ok {
			for _, tag := range tags {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					if name, ok := tagMap["name"].(string); ok {
						info.Genres = append(info.Genres, name)
					}
				}
			}
		}

		// Handle studio
		if brandMeta, ok := result["brandMeta"].(map[string]interface{}); ok {
			if name, ok := brandMeta["name"].(string); ok {
				info.Maker = name
			}
		}

		info.Homepage = fmt.Sprintf("%s/scene/%s/%s", a.baseURL, info.ID, slugify(info.Title))
	})

	err = c.Request("GET", apiURL, nil, nil, headers)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ParseMovieIDFromURL implements MovieProvider.ParseMovieIDFromURL
func (a *Aylo) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Extract ID from path like /scene/12345/some-title
	parts := strings.Split(u.Path, "/")
	for i, part := range parts {
		if part == "scene" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return path.Base(u.Path), nil
}

// GetMovieInfoByURL implements MovieProvider.GetMovieInfoByURL
func (a *Aylo) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := a.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	return a.GetMovieInfoByID(id)
}

// NormalizeMovieKeyword implements MovieSearcher.NormalizeMovieKeyword
func (a *Aylo) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return number.NormalizeMovieKeyword(keyword, "")
}

// SearchMovie implements MovieSearcher.SearchMovie
func (a *Aylo) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	// Ensure token is valid before making request
	if err := a.getTokenWithAutoRefresh(); err != nil {
		return nil, err
	}

	query := url.QueryEscape(a.NormalizeMovieKeyword(keyword))
	apiURL := fmt.Sprintf("%s/%s%s", APIBaseURL, APIVersion, fmt.Sprintf(SearchEndpoint, query))

	// Create request with headers
	headers := http.Header{}
	headers.Set("Instance", a.token)
	headers.Set("User-Agent", userAgent)
	headers.Set("Origin", a.baseURL)
	headers.Set("Referer", a.baseURL)

	// Create a new collector for this request
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		var apiResp struct {
			Result []map[string]interface{} `json:"result"`
		}

		if err = json.Unmarshal(r.Body, &apiResp); err != nil {
			return
		}

		for _, item := range apiResp.Result {
			// Only include scenes
			if item["type"] != "scene" {
				continue
			}

			id := fmt.Sprintf("%.0f", item["id"])
			title := fmt.Sprintf("%v", item["title"])

			result := &model.MovieSearchResult{
				ID:       id,
				Number:   id,
				Title:    title,
				Provider: a.Name(),
			}

			// Handle date
			if dateStr, ok := item["dateReleased"].(string); ok {
				if date, err := time.Parse(time.RFC3339, dateStr); err == nil {
					result.ReleaseDate = datatypes.Date(date)
				}
			}

			// Handle images
			if images, ok := item["images"].(map[string]interface{}); ok {
				if poster, ok := images["poster"].(map[string]interface{}); ok {
					if first, ok := poster["0"].(map[string]interface{}); ok {
						if xx, ok := first["xx"].(map[string]interface{}); ok {
							if url, ok := xx["url"].(string); ok {
								result.ThumbURL = strings.Split(url, "/m=")[0] + "/m=COV_400x600"
								result.CoverURL = strings.Split(url, "/m=")[0] + "/m=COV_400x600"
							}
						}
					}
				}
			}

			result.Homepage = fmt.Sprintf("%s/scene/%s/%s", a.baseURL, result.ID, slugify(title))

			results = append(results, result)
		}
	})

	err = c.Request("GET", apiURL, nil, nil, headers)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (a *Aylo) getToken() error {
	// Create a new collector for token requests
	c := a.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		for _, cookie := range r.Headers.Values("Set-Cookie") {
			if strings.Contains(cookie, "instance_token=") {
				// Extract token value
				parts := strings.Split(cookie, "instance_token=")
				if len(parts) > 1 {
					tokenPart := parts[1]
					if idx := strings.Index(tokenPart, ";"); idx > 0 {
						a.token = tokenPart[:idx]
					} else {
						a.token = tokenPart
					}
					a.tokenExpire = time.Now().AddDate(0, 1, 0)
					break
				}
			}
		}
	})

	if err := c.Request(http.MethodHead, a.baseURL, nil, nil, http.Header{"User-Agent": {userAgent}}); err != nil {
		return err
	}
	return nil
}

// getTokenWithAutoRefresh 获取 token，如果 token 已过期则自动刷新
func (a *Aylo) getTokenWithAutoRefresh() error {
	if a.token == "" || time.Now().After(a.tokenExpire) {
		return a.getToken()
	}
	return nil
}

// slugify creates a URL-friendly slug from a string
func slugify(text string) string {
	// Simple slugify implementation
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, ";", "")
	slug = strings.ReplaceAll(slug, "!", "")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "/", "")
	slug = strings.ReplaceAll(slug, "\\", "")
	slug = strings.ReplaceAll(slug, "(", "")
	slug = strings.ReplaceAll(slug, ")", "")

	// Remove any remaining non-alphanumeric characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
