package fc2hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2/fc2util"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*FC2HUB)(nil)
	_ provider.MovieSearcher = (*FC2HUB)(nil)
)

const (
	Name     = "fc2hub"
	Priority = 1000 - 1
)

const (
	baseURL   = "https://javten.com/"
	movieURL  = "https://javten.com/video/%s/id%s/%s"
	searchURL = "https://javten.com/search?kw=%s"
)

type FC2HUB struct {
	*scraper.Scraper
}

func New() *FC2HUB {
	return &FC2HUB{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (fc2hub *FC2HUB) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	ss := strings.SplitN(id, "-", 2)
	if len(ss) != 2 {
		return nil, provider.ErrInvalidID
	}
	const padding = "%20" // use padding to fix weird colly trailing path issue.
	return fc2hub.GetMovieInfoByURL(fmt.Sprintf(movieURL, ss[0], ss[1], padding))
}

func (fc2hub *FC2HUB) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if ss := regexp.MustCompile(`/video/(\d+)/id(\d+)`).FindStringSubmatch(homepage.Path); len(ss) == 3 {
		return fmt.Sprintf("%s-%s", ss[1], ss[2]), nil
	}
	return "", provider.ErrInvalidURL
}

func (fc2hub *FC2HUB) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := fc2hub.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id, // Dual-ID (id+number)
		Provider:      fc2hub.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := fc2hub.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="content"]/div/div[2]/div[1]/div[1]/div[2]/h1`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//*[@id="content"]/div/div[2]/div[1]/div[1]/div[2]/div[2]/div`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Number
	c.OnXML(`//*[@id="content"]/div/div[2]/div[1]/div[1]/div[2]/div[1]/div[2]/h1`, func(e *colly.XMLElement) {
		if num := fc2util.ParseNumber(strings.TrimSpace(e.Text)); num != "" {
			info.Number = fmt.Sprintf("FC2-%s", num)
		}
	})

	// Genres
	c.OnXML(`//*[@id="content"]/div/div[2]/div[1]/div[1]/div[2]/p/a`, func(e *colly.XMLElement) {
		if genre := strings.TrimSpace(e.Text); genre != "" {
			info.Genres = append(info.Genres, genre)
		}
	})

	// Maker
	c.OnXML(`//*[@id="content"]/div/div[2]/div[1]/div[3]/div/div[2]/div/div[2]`, func(e *colly.XMLElement) {
		// info.Maker = strings.TrimSpace(strings.Split(e.Text, "\n")[0])
		for n := e.DOM.(*html.Node).FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				info.Maker = strings.TrimSpace(n.Data)
				break
			}
		}
	})

	// Preview Images
	c.OnXML(`//*[@id="content"]/div/div[2]/div[1]/div[2]/div[3]/div/div//a[@data-fancybox="gallery"]`, func(e *colly.XMLElement) {
		if href := e.Attr("href"); href != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(href))
		}
	})

	// Fields
	c.OnXML(`/html/head/script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Type string `json:"@type"`
			// `Movie`
			Name          string   `json:"name"`
			Description   string   `json:"description"`
			Image         string   `json:"image"`
			Identifier    []string `json:"identifier"`
			DatePublished string   `json:"datePublished"`
			Duration      string   `json:"duration"`
			Actor         []string `json:"actor"`
			Genre         []string `json:"genre"`
			Director      string   `json:"director"`
			// `CreativeWorkSeries`
			AggregateRating struct {
				BestRating  float64 `json:"bestRating"`
				WorstRating float64 `json:"worstRating"`
				RatingCount int     `json:"ratingCount"`
				RatingValue float64 `json:"ratingValue"`
			}
			// `WebPage`
			URL string `json:"url"`
		}{}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			switch data.Type {
			case "Movie":
				if data.Name != "" {
					info.Title = data.Name
				}
				if info.Summary == "" {
					info.Summary = data.Description
				}
				if data.Director != "" {
					// Use director as maker.
					info.Maker = data.Director
				}
				if len(info.Genres) == 0 {
					info.Genres = removeEmpty(data.Genre)
				}
				if len(data.Actor) > 0 {
					info.Actors = removeEmpty(data.Actor)
				}
				for _, identifier := range data.Identifier {
					if num := fc2util.ParseNumber(identifier); num != "" {
						info.Number = fmt.Sprintf("FC2-%s", num)
						break
					}
				}
				info.CoverURL = data.Image
				info.ReleaseDate = parser.ParseDate(data.DatePublished)
				info.Runtime = parser.ParseRuntime(data.Duration)
			case "CreativeWorkSeries":
				// Average rating score.
				info.Score = data.AggregateRating.RatingValue
			case "WebPage":
				//if data.URL != "" {
				//	// Update homepage URL.
				//	info.Homepage = data.URL
				//}
			}
		}
	})

	// Cover (fallback)
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" && len(info.PreviewImages) > 0 {
			info.CoverURL = info.PreviewImages[0]
		}
		// cover as thumb image.
		info.ThumbURL = info.CoverURL
	})

	// Homepage (update)
	c.OnScraped(func(_ *colly.Response) {
		if info.ID != "" && len(info.Number) > 4 && info.Title != "" {
			info.Homepage = fmt.Sprintf(movieURL,
				strings.SplitN(info.ID, "-", 2)[0],
				info.Number[4:],
				url.PathEscape(info.Title))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (fc2hub *FC2HUB) NormalizeMovieKeyword(keyword string) string {
	return fc2util.ParseNumber(keyword)
}

func (fc2hub *FC2HUB) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := fc2hub.ClonedCollector()
	c.ParseHTTPErrorResponse = true
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	c.OnResponse(func(r *colly.Response) {
		var loc *url.URL
		if loc, err = url.Parse(r.Request.AbsoluteURL(r.Headers.Get("Location"))); err != nil {
			return
		}
		if regexp.MustCompile(`/video/\d+/id\d+`).MatchString(loc.Path) {
			var info *model.MovieInfo
			if info, err = fc2hub.GetMovieInfoByURL(loc.String()); err != nil {
				return
			}
			results = append(results, info.ToSearchResult())
		}
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func removeEmpty(in []string) (out []string) {
	if len(in) == 0 {
		return in
	}
	out = make([]string, 0, len(in))
	for _, elem := range in {
		if strings.TrimSpace(elem) != "" {
			out = append(out, elem)
		}
	}
	return
}

func init() {
	provider.Register(Name, New)
}
