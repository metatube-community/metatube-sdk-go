package h0930

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"

	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*H0930)(nil)

const (
	Name     = "H0930"
	Priority = 1000
)

const (
	baseURL  = "https://www.h0930.com/"
	movieURL = "https://www.h0930.com/moviepages/%s/index.html"
)

type H0930 struct {
	*scraper.Scraper
}

func New() *H0930 {
	return &H0930{scraper.NewDefaultScraper(Name, baseURL, Priority, scraper.WithDetectCharset())}
}

func (h *H0930) NormalizeID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:h0930[-_])?([a-z\d]+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return strings.ToLower(ss[1])
	}
	return ""
}

func (h *H0930) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return h.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (h *H0930) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(path.Dir(homepage.Path)), nil
}

func (h *H0930) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := h.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("H0930-%s", strings.ToUpper(id)),
		Provider:      h.Name(),
		Homepage:      rawURL,
		Maker:         "エッチな0930",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := h.ClonedCollector()

	// JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Name  string `json:"name"`
			Image string `json:"image"`
			Actor struct {
				Name  string `json:"name"`
				Image string `json:"image"`
			} `json:"actor"`
			Description   string `json:"description"`
			Duration      string `json:"duration"`
			DateCreated   string `json:"dateCreated"`
			ReleasedEvent struct {
				StartDate string `json:"startDate"`
				Location  struct {
					Name string `json:"name"`
				} `json:"location"`
			} `json:"releasedEvent"`
			Video struct {
				Thumbnail string `json:"thumbnail"`
				Duration  string `json:"duration"`
				Actor     string `json:"actor"`
				Provider  string `json:"provider"`
			} `json:"video"`
			AggregateRating struct {
				RatingValue string `json:"ratingValue"`
			} `json:"aggregateRating"`
		}{}
		if json.Unmarshal([]byte(strings.ReplaceAll(e.Text, "\n", "")), &data) == nil {

			info.Title = data.Name
			info.Summary = data.Description
			if data.Image != "" {
				info.CoverURL = e.Request.AbsoluteURL(data.Image)
				info.ThumbURL = info.CoverURL /* use cover as thumb */
			}
			info.ReleaseDate = parser.ParseDate(data.ReleasedEvent.StartDate)
			info.Runtime = parser.ParseRuntime(data.Video.Duration)
			info.Score = parser.ParseScore(data.AggregateRating.RatingValue)
			if data.Video.Provider != "" {
				info.Maker = data.Video.Provider
			}
			if data.Video.Actor != "" {
				info.Actors = []string{data.Video.Actor}
			}
		}
	})

	// Title
	c.OnXML(`//*[@id="moviePlay"]//div[@class="moviePlay_title"]/h1/span`, func(e *colly.XMLElement) {
		if title := strings.TrimSpace(e.Text); title != "" {
			info.Title = title
		}
	})

	// Thumb+Cover+Preview Video
	c.OnXML(`//*[@id="movieContent"]`, func(e *colly.XMLElement) {
		if poster := e.Attr("poster"); poster != "" {
			info.CoverURL = e.Request.AbsoluteURL(poster)
			info.ThumbURL = info.CoverURL /* use cover as thumb */
		}
		if src := e.ChildAttr(`./source`, "src"); src != "" {
			info.PreviewVideoURL = e.Request.AbsoluteURL(src)
		}
	})

	// Preview Images
	c.OnXML(`//*[@id="movieGallery"]//script[@type="text/javascript"]`, func(e *colly.XMLElement) {
		if ss := regexp.MustCompile(`href="(.+?)"`).FindAllStringSubmatch(e.Text, -1); len(ss) > 0 {
			for _, ss := range ss {
				if !strings.Contains(ss[1], "member") {
					info.PreviewImages = append(info.PreviewImages, ss[1])
				}
			}
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
