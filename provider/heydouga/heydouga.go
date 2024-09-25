package heydouga

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*HeyDouga)(nil)

const (
	Name     = "HeyDouga"
	Priority = 1000
)

const (
	baseURL     = "https://www.heydouga.com/"
	movieURL    = "https://www.heydouga.com/moviepages/%s/%s/index.html"
	movieTagURL = "https://www.heydouga.com/get_movie_tag_all/"
)

type HeyDouga struct {
	*scraper.Scraper
}

func New() *HeyDouga {
	return &HeyDouga{scraper.NewDefaultScraper(Name, baseURL, Priority)}
}

func (hey *HeyDouga) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:heydouga[-_])?(\d{4}-[a-z\d]+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (hey *HeyDouga) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	if ss := strings.SplitN(id, "-", 2); len(ss) == 2 {
		return hey.GetMovieInfoByURL(fmt.Sprintf(movieURL, ss[0], ss[1]))
	}
	return nil, provider.ErrInvalidID
}

func (hey *HeyDouga) ParseMovieIDFromURL(rawURL string) (id string, err error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if ss := regexp.MustCompile(`/(\d+)/(\d+)/index\.html`).
		FindStringSubmatch(homepage.Path); len(ss) == 3 {
		id = fmt.Sprintf("%s-%s", ss[1], ss[2])
	}
	return
}

func (hey *HeyDouga) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := hey.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("HEYDOUGA-%s", id),
		Provider:      hey.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := hey.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="title-bg"]/h1`, func(e *colly.XMLElement) {
		for n := e.DOM.(*html.Node).FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				if title := strings.TrimSpace(n.Data); title != "" {
					info.Title = strings.TrimSpace(title)
					return
				}
			}
		}
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//*[@id="movie-detail-mobile"]/div/p[1]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Cover
	c.OnXML(`//section[@class="movie-player"]//script`, func(e *colly.XMLElement) {
		if ss := regexp.MustCompile(`(?i)player_poster\s*=\s*'(\/.+?)';`).FindStringSubmatch(e.Text); len(ss) == 2 {
			info.CoverURL = e.Request.AbsoluteURL(ss[1])
			info.ThumbURL = info.CoverURL
		}
	})

	// Fields
	c.OnXML(`//*[@id="movie-info"]/ul/li`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span[1]`) {
		case "Release Date：":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//span[2]`))
		case "Actor：":
			// heydouga's actor info is sticky, but whatever...
			info.Actors = strings.Fields(e.ChildText(`.//span[2]`))
		case "Provider：":
			if info.Maker = strings.TrimSpace(e.ChildText(`.//span[2]/a[1]`)); info.Maker == "" /* fallback */ {
				info.Maker = strings.TrimSpace(e.ChildText(`.//span[2]`))
			}
		case "Movie Duration：":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//span[2]`))
		case "ファイル容量：", "画面サイズ：":
			// skip, do nothing
		}
	})

	// API Query
	c.OnScraped(func(r *colly.Response) {
		var wg sync.WaitGroup
		body := string(r.Body)

		// Preview Video
		{
			if ss := regexp.MustCompile(`source\s*=\s*'(.+\.m3.*?)';`).
				FindStringSubmatch(body); len(ss) == 2 {
				info.PreviewVideoHLSURL = r.Request.AbsoluteURL(ss[1])
			}
		}

		// Score
		wg.Add(1)
		go func() {
			defer wg.Done()
			var ratingURL string
			if ss := regexp.MustCompile(`url_get_movie_rating\s*=\s*"(.+?)";`).
				FindStringSubmatch(body); len(ss) == 2 {
				ratingURL = ss[1]
			}
			if ratingURL != "" {
				d := c.Clone()
				d.OnResponse(func(r *colly.Response) {
					data := struct {
						MovieRatingAverage string `json:"movie_rating_average"`
						MovieRatingCount   string `json:"movie_rating_count"`
					}{}
					if json.Unmarshal(r.Body, &data) == nil {
						info.Score = parser.ParseScore(data.MovieRatingAverage)
					}
				})
				d.Visit(r.Request.AbsoluteURL(ratingURL))
			}
		}()

		// Genres
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				providerID string
				movieSeq   string
			)
			if ss := regexp.MustCompile(`provider_id\s*=\s*(\d+);`).
				FindStringSubmatch(body); len(ss) == 2 {
				providerID = ss[1]
			}
			if ss := regexp.MustCompile(`data\s*:\s*\{\s*movie_seq\s*:\s*(\d+),`).
				FindStringSubmatch(body); len(ss) == 2 {
				movieSeq = ss[1]
			}
			if providerID != "" && movieSeq != "" {
				d := c.Clone()
				d.OnResponse(func(r *colly.Response) {
					data := struct {
						Tag []struct {
							TagID      int    `json:"tag_id"`
							TagName    string `json:"tag_name"`
							TagNameURI string `json:"tag_name_uri"`
						} `json:"tag"`
					}{}
					if json.Unmarshal(r.Body, &data) == nil {
						for _, tag := range data.Tag {
							info.Genres = append(info.Genres, tag.TagName)
						}
					}
				})
				d.Post(movieTagURL, map[string]string{
					"movie_seq":   movieSeq,
					"provider_id": providerID,
				})
			}
		}()

		wg.Wait()
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
