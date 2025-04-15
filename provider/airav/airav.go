// Deprecated: This provider is no longer supported.
package airav

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*AirAV)(nil)
	_ provider.MovieSearcher = (*AirAV)(nil)
)

const (
	Name     = "AIRAV"
	Priority = 1000 - 6 // unofficial provider gets lower priority.
)

const (
	baseURL      = "https://www.airav.wiki/"
	movieURL     = "https://www.airav.wiki/video/%s"
	movieAPIURL  = "https://www.airav.wiki/api/video/barcode/%s?lng=jp"
	searchAPIURL = "https://www.airav.wiki/api/video/list?search=%s&lng=jp"
	videoAPIURL  = "https://www.airav.wiki/api/video/getVideoMedia?barcode=%s&vid=%s"
)

type AirAV struct {
	*scraper.Scraper
}

func New() *AirAV {
	return &AirAV{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		scraper.WithHeaders(map[string]string{
			"Origin":  baseURL,
			"Referer": baseURL,
		}),
	)}
}

func (air *AirAV) NormalizeMovieID(id string) string { return strings.ToUpper(id) }

func (air *AirAV) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return air.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (air *AirAV) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return air.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (air *AirAV) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := air.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		Provider:      air.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := air.ClonedCollector()

	// JSON
	c.OnResponse(func(r *colly.Response) {
		data := struct {
			Result struct {
				Actors []struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					NameJP string `json:"name_jp"`
				} `json:"actors"`
				Barcode     string `json:"barcode"`
				Description string `json:"description"`
				Factories   []struct {
					Name string `json:"name"`
				} `json:"factories"`
				Images      []string `json:"images"`
				ImgURL      string   `json:"img_url"`
				Name        string   `json:"name"`
				PublishDate string   `json:"publish_date"`
				Tags        []struct {
					Name string `json:"name"`
				} `json:"tags"`
				VID      string `json:"vid"`
				VideoURL struct {
					URLCDN string `json:"url_cdn"`
				} `json:"video_url"`
				View int `json:"view"`
			} `json:"result"`
			Count  int    `json:"count"`
			Status string `json:"status"`
		}{}
		if err = json.Unmarshal(r.Body, &data); err == nil && data.Count > 0 {
			info.ID = data.Result.Barcode
			info.Number = ParseNumber(data.Result.Barcode)
			info.Title = data.Result.Name
			info.Summary = data.Result.Description
			info.ThumbURL = data.Result.ImgURL
			info.CoverURL = data.Result.ImgURL
			info.PreviewImages = data.Result.Images
			info.ReleaseDate = parser.ParseDate(data.Result.PublishDate)
			if len(data.Result.Factories) > 0 {
				info.Maker = data.Result.Factories[0].Name
			}
			for _, tag := range data.Result.Tags {
				info.Genres = append(info.Genres, tag.Name)
			}
			for _, actor := range data.Result.Actors {
				if actor.NameJP != "" {
					info.Actors = append(info.Actors, actor.NameJP)
				} else if actor.Name != "" {
					info.Actors = append(info.Actors, actor.Name)
				}
			}
			if data.Result.VideoURL.URLCDN != "" {
				info.PreviewVideoURL = data.Result.VideoURL.URLCDN
			} else {
				d := c.Clone()
				d.OnResponse(func(r *colly.Response) {
					videoData := struct {
						Data struct {
							Msg       string `json:"msg"`
							URL       string `json:"url"`
							URLCDN    string `json:"url_cdn"`
							URLHLS    string `json:"url_hls"`
							URLHLSCDN string `json:"url_hls_cdn"`
						} `json:"data"`
						Status string `json:"status"`
					}{}
					if json.Unmarshal(r.Body, &videoData) == nil {
						for _, videoURL := range []string{
							videoData.Data.URL, videoData.Data.URLCDN,
						} {
							if videoURL != "" {
								info.PreviewVideoURL = videoURL
								break
							}
						}
						for _, videoURL := range []string{
							videoData.Data.URLHLS, videoData.Data.URLHLSCDN,
						} {
							if videoURL != "" {
								info.PreviewVideoHLSURL = videoURL
								break
							}
						}
					}
				})
				d.Visit(fmt.Sprintf(videoAPIURL, data.Result.Barcode, data.Result.VID))
			}
		}
	})

	if vErr := c.Visit(fmt.Sprintf(movieAPIURL, id)); vErr != nil {
		err = vErr
	}
	return
}

func (air *AirAV) NormalizeMovieKeyword(keyword string) string {
	if ss := regexp.MustCompile(`^(?i)FC2-.*?(\d+)$`).FindStringSubmatch(keyword); len(ss) == 2 {
		return fmt.Sprintf("FC2-PPV-%s", ss[1])
	}
	if number.IsSpecial(keyword) {
		return "" // Deprecate other searches if any.
	}
	return strings.ToUpper(keyword)
}

func (air *AirAV) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := air.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			Count  int `json:"count"`
			Offset int `json:"offset"`
			Result []struct {
				Barcode string `json:"barcode"`
				ImgURL  string `json:"img_url"`
				Name    string `json:"name"`
				VID     string `json:"vid"`
			}
			Status string `json:"status"`
		}{}
		if json.Unmarshal(r.Body, &data) == nil {
			for _, result := range data.Result {
				results = append(results, &model.MovieSearchResult{
					ID:       result.Barcode,
					Number:   ParseNumber(result.Barcode),
					Title:    result.Name,
					Provider: air.Name(),
					Homepage: fmt.Sprintf(movieURL, result.Barcode),
					ThumbURL: result.ImgURL,
					CoverURL: result.ImgURL,
				})
			}
		}
	})

	err = c.Visit(fmt.Sprintf(searchAPIURL, url.QueryEscape(keyword)))
	return
}

// ParseNumber parses barcode to standard movie number.
func ParseNumber(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "FC2-PPV-", "FC2-") // Use `FC2` directly
	return s
}

func init() {
	provider.Register(Name, New)
}
