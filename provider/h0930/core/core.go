package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

type Core struct {
	*scraper.Scraper

	// URLs
	BaseURL  string
	MovieURL string

	// Values
	DefaultPriority float64
	DefaultName     string
	DefaultMaker    string
}

func (core *Core) Init() *Core {
	core.Scraper = scraper.NewDefaultScraper(
		core.DefaultName,
		core.BaseURL,
		core.DefaultPriority,
		language.Japanese,
		scraper.WithDetectCharset())
	return core
}

func (core *Core) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return core.GetMovieInfoByURL(fmt.Sprintf(core.MovieURL, id))
}

func (core *Core) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(path.Dir(homepage.Path)), nil
}

func (core *Core) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := core.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        strings.ToLower(fmt.Sprintf("%s-%s", core.DefaultName, id)),
		Provider:      core.Name(),
		Homepage:      rawURL,
		Maker:         core.DefaultMaker,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := core.ClonedCollector()

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

	// Fields
	c.OnXML(`//*[@id="movieInfo"]//section/dl`, func(e *colly.XMLElement) {
		for i, dt := range e.ChildTexts(`.//dt`) {
			dd := fmt.Sprintf(`.//dd[%d]`, i+1)
			switch dt {
			case "年齢":
			case "身長":
			case "3サイズ":
			case "タイプ":
			case "動画":
				if info.Runtime == 0 {
					info.Runtime = parser.ParseRuntime(e.ChildText(dd))
				}
			case "公開日":
				if time.Time(info.ReleaseDate).IsZero() {
					info.ReleaseDate = parser.ParseDate(e.ChildText(dd))
				}
			case "プレイ内容":
				// info.Genres = strings.Fields(e.ChildText(dd))
				for _, genre := range strings.Split(e.ChildText(dd), "\u00a0") {
					if genre := strings.TrimSpace(genre); genre != "" {
						info.Genres = append(info.Genres, genre)
					}
				}
			}
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
