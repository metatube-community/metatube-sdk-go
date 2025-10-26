package fit18

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fit18/internal/graphql"
)

const (
	Fit18Name     = "Fit18"
	Fit18URL      = "https://fit18.com"
	APIURL        = "https://%s.team18media.app/graphql"
	Fit18APIKey   = "77cd9282-9d81-4ba8-8868-ca9125c76991"
	Thicc18Name   = "Thicc18"
	Thicc18URL    = "https://thicc18.com"
	Thicc18APIKey = "0e36c7e9-8cb7-4fa1-9454-adbc2bad15f0"
	Priority      = 1000
)

var (
	_ provider.MovieProvider = (*Fit18)(nil)
	_ provider.MovieSearcher = (*Fit18)(nil)
)

type Fit18 struct {
	name     string
	baseURL  *url.URL
	graphQL  *graphql.Client
	priority float64
	language language.Tag
}

func New() *Fit18 {
	baseURL, _ := url.Parse(Fit18URL)
	apiURL := fmt.Sprintf(APIURL, strings.ToLower(Fit18Name))
	return &Fit18{
		name:     Fit18Name,
		baseURL:  baseURL,
		graphQL:  graphql.NewClient(apiURL, Fit18APIKey),
		priority: Priority,
		language: language.English,
	}
}

func NewThicc18() *Fit18 {
	baseURL, _ := url.Parse(Thicc18URL)
	apiURL := fmt.Sprintf(APIURL, strings.ToLower(Thicc18Name))
	return &Fit18{
		name:     Thicc18Name,
		baseURL:  baseURL,
		graphQL:  graphql.NewClient(apiURL, Thicc18APIKey),
		priority: Priority,
		language: language.English,
	}
}

func (f *Fit18) Name() string {
	return f.name
}

func (f *Fit18) URL() *url.URL {
	return f.baseURL
}

func (f *Fit18) Priority() float64 {
	return f.priority
}

func (f *Fit18) SetPriority(priority float64) {
	f.priority = priority
}

func (f *Fit18) Language() language.Tag {
	return f.language
}

func (f *Fit18) NormalizeMovieID(id string) string {
	return strings.ToLower(id)
}

func (f *Fit18) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Check if the URL belongs to our domain
	if homepage.Host != f.baseURL.Host {
		return "", fmt.Errorf("URL does not belong to %s: %s", f.baseURL.Host, rawURL)
	}

	// Check if the URL path starts with "/videos/"
	if !strings.HasPrefix(homepage.Path, "/videos/") {
		return "", fmt.Errorf("invalid URL path: %s", homepage.Path)
	}

	// Extract the ID from the path
	id := strings.TrimPrefix(homepage.Path, "/videos/")
	// Remove trailing slash if present
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		return "", fmt.Errorf("empty ID in URL: %s", rawURL)
	}

	// Validate ID format (should contain at least one colon)
	if !strings.Contains(id, ":") {
		return "", fmt.Errorf("invalid ID format, missing colon: %s", id)
	}

	return id, nil
}

func (f *Fit18) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return f.getMovieInfoFromHTML(id)
}

func (f *Fit18) getMovieInfoFromHTML(id string) (info *model.MovieInfo, err error) {
	// Construct the URL for the video page
	videoURL := fmt.Sprintf("%s/videos/%s", f.baseURL.String(), url.QueryEscape(id))

	// Create a new collector
	c := colly.NewCollector(
		colly.IgnoreRobotsTxt(),
	)

	// Create movie info object
	info = &model.MovieInfo{
		ID:       id,
		Number:   id,
		Provider: f.Name(),
		Homepage: f.baseURL.String(),
		Actors:   []string{},
		Genres:   []string{},
		Maker:    f.Name(),
		Label:    f.Name(),
	}

	// Extract data from the script tag containing JSON
	c.OnHTML("script", func(e *colly.HTMLElement) {
		// Look for the script tag with window.__INITIAL__DATA__
		scriptContent := e.Text
		if strings.Contains(scriptContent, "window.__INITIAL__DATA__") {
			// Extract JSON data from the script content
			start := strings.Index(scriptContent, "JSON.parse(\"")
			if start == -1 {
				return
			}

			// Extract the JSON string
			jsonStart := start + 12 // Length of "JSON.parse(""
			// Find the end of the JSON string (look for the closing quote and parenthesis)
			jsonEnd := strings.Index(scriptContent[jsonStart:], "\")")
			if jsonEnd == -1 {
				return
			}

			jsonStr := scriptContent[jsonStart : jsonStart+jsonEnd]
			// Unescape the JSON string
			jsonStr = strings.ReplaceAll(jsonStr, "\\\"", "\"")
			jsonStr = strings.ReplaceAll(jsonStr, "\\\\", "\\")

			// Parse the JSON data
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
				return
			}

			// Extract page data
			if page, ok := data["page"].(map[string]interface{}); ok {
				if current, ok := page["current"].(map[string]interface{}); ok {
					if currentType, ok := current["$page"].(string); ok && currentType == "video" {
						if video, ok := current["video"].(map[string]interface{}); ok {
							// Extract title
							if title, ok := video["title"].(string); ok {
								info.Title = title
							}

							// Extract duration
							if duration, ok := video["duration"].(int); ok {
								info.Runtime = duration
							}

							// Extract description
							if description, ok := video["description"].(map[string]interface{}); ok {
								if long, ok := description["long"].(string); ok && long != "" {
									info.Summary = long
								} else if short, ok := description["short"].(string); ok {
									info.Summary = short
								}
							}

							// Extract talent/actors
							if talents, ok := video["talent"].([]interface{}); ok {
								for _, t := range talents {
									if talent, ok := t.(map[string]interface{}); ok {
										if talentType, ok := talent["type"].(string); ok && talentType == "MODEL" {
											if talentInfo, ok := talent["talent"].(map[string]interface{}); ok {
												if name, ok := talentInfo["name"].(string); ok {
													info.Actors = append(info.Actors, name)
												}
											}
										}
									}
								}
							}

							// Extract poster/thumbnail
							if poster, ok := video["poster"].(map[string]interface{}); ok {
								if thumbURL, ok := poster["x1"].(string); ok {
									info.ThumbURL = thumbURL
									info.CoverURL = thumbURL
									info.BigThumbURL = thumbURL
									info.BigCoverURL = thumbURL
								}
							}
						}
					}
				}
			}
		}
	})

	// Visit the page
	if err := c.Visit(videoURL); err != nil {
		return nil, fmt.Errorf("failed to scrape movie info from HTML: %w", err)
	}

	// Check if we got the essential information
	if info.Title == "" {
		return nil, fmt.Errorf("failed to extract movie info from HTML")
	}

	return info, nil
}

func (f *Fit18) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := f.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse movie ID from URL: %w", err)
	}
	return f.GetMovieInfoByID(id)
}

func (f *Fit18) extractNumberFromID(id string) string {
	// Extract number from ID (format: talentId:sceneNumber)
	parts := strings.Split(id, ":")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return id
}

//func (f *Fit18) getImageURL(talentID, sceneNumber string) string {
//	// Construct image path
//	imagePath := fmt.Sprintf("/members/models/%s/scenes/%s/videothumb.jpg", talentID, sceneNumber)
//
//	// Call GraphQL API to get asset URL
//	resp, err := f.graphQL.BatchFindAsset([]string{imagePath})
//	if err != nil {
//		return ""
//	}
//
//	// Check if response is valid
//	if resp == nil || len(resp.Asset.Batch.Result) == 0 {
//		return ""
//	}
//
//	// Return image URL
//	asset := resp.Asset.Batch.Result[0]
//	if asset.Serve.URI != "" {
//		return asset.Serve.URI
//	}
//
//	return ""
//}

func (f *Fit18) NormalizeMovieKeyword(keyword string) string {
	return strings.TrimSpace(keyword)
}

func (f *Fit18) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	defer func() {
		f.sortMovieSearchResults(keyword, results)
	}()

	searchKeyword := f.NormalizeMovieKeyword(keyword)

	resp, err := f.graphQL.Search(searchKeyword)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}

	// Check if the response is valid
	if resp == nil {
		return nil, fmt.Errorf("received nil response from search API")
	}

	// Collect all image paths for batch asset query
	var allImagePaths []string
	for _, item := range resp.Search.Search.Result {
		if item.Type == "VIDEO" {
			allImagePaths = append(allImagePaths, item.Images...)
		}
	}

	// Get image URLs through batch asset query
	imageURLMap := make(map[string]string)
	if len(allImagePaths) > 0 {
		assetResp, err := f.graphQL.BatchFindAsset(allImagePaths)
		if err != nil {
			fmt.Printf("Warning: failed to get asset URLs: %v\n", err)
		} else {
			for _, asset := range assetResp.Asset.Batch.Result {
				imageURLMap[asset.Path] = asset.Serve.URI
			}
		}
	}

	for _, item := range resp.Search.Search.Result {
		if item.Type != "VIDEO" {
			continue // Only process video items
		}

		// Get thumbnail URL
		var thumbURL string
		if len(item.Images) > 0 {
			thumbPath := item.Images[0]
			if u, ok := imageURLMap[thumbPath]; ok {
				thumbURL = u
			} else {
				// Fallback to constructing URL
				thumbURL = fmt.Sprintf("%s%s", f.baseURL.String(), thumbPath)
			}
		}

		results = append(results, &model.MovieSearchResult{
			ID:       item.ItemID,
			Number:   item.ItemID,
			Title:    item.Name,
			Provider: f.Name(),
			Homepage: fmt.Sprintf("%s/videos/%s", f.baseURL.String(), item.ItemID),
			ThumbURL: thumbURL,
			CoverURL: thumbURL,
		})
	}

	return results, nil
}

func (f *Fit18) sortMovieSearchResults(keyword string, results []*model.MovieSearchResult) {
	if len(results) == 0 {
		return
	}

	// Sort results based on similarity to the keyword
	sort.SliceStable(results, func(i, j int) bool {
		return comparer.Compare(results[i].Title, keyword) > comparer.Compare(results[j].Title, keyword)
	})
}

func init() {
	provider.Register(Fit18Name, New)
	provider.Register(Thicc18Name, NewThicc18)
}
