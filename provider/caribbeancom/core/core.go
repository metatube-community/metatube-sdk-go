package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

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
	DefaultPriority int
	DefaultName     string
	DefaultMaker    string
}

func (core *Core) Init() *Core {
	core.Scraper = scraper.NewDefaultScraper(
		core.DefaultName,
		core.BaseURL,
		core.DefaultPriority,
		scraper.WithDetectCharset())
	return core
}

func (core *Core) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return core.GetMovieInfoByURL(fmt.Sprintf(core.MovieURL, id))
}

func (core *Core) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(path.Dir(homepage.Path)), nil
}

func (core *Core) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := core.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        id,
		Provider:      core.Name(),
		Homepage:      rawURL,
		Maker:         core.DefaultMaker,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := core.ClonedCollector()

	// Title
	c.OnXML(`//h1[@itemprop="name"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//p[@itemprop="description"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Title+Summary (Fallback)
	c.OnXML(`//div[@id="moviepages"]`, func(e *colly.XMLElement) {
		if info.Title == "" {
			info.Title = strings.TrimSpace(e.ChildText(`.//h1[1]`))
		}
		if info.Summary == "" {
			info.Summary = strings.TrimSpace(e.ChildText(`.//p[1]`))
		}
	})

	// Fields
	c.OnXML(`//*[@id="moviepages"]//li`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span[1]`) {
		case "出演":
			var actors []string
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//span[2]`), &actors)
			for _, actor := range actors {
				if actor := strings.Trim(actor, "-"); actor != "" {
					info.Actors = append(info.Actors, actor)
				}
			}
		case "配信日", "販売日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//span[2]`))
		case "再生時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//span[2]`))
		case "シリーズ":
			info.Series = e.ChildText(`.//span[2]/a[1]`)
		case "スタジオ":
			info.Maker /* studio */ = e.ChildText(`.//span[2]/a[1]`)
		case "タグ":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//span[2]`),
				(*[]string)(&info.Genres))
		case "ユーザー評価":
			info.Score = float64(utf8.RuneCountInString(
				strings.TrimSpace(e.ChildText(`.//span[2]`))))
		}
	})

	// Thumb+Cover+Video
	c.OnXML(`//script`, func(e *colly.XMLElement) {
		if re := regexp.MustCompile(`emimg\s*=\s*'(.+?)';`); re.MatchString(e.Text) {
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 2 {
				info.ThumbURL = e.Request.AbsoluteURL(ss[1])
				info.CoverURL = info.ThumbURL /* use thumb as cover */
			}
		} else if re = regexp.MustCompile(`posterImage\s*=\s*'(.+?)'\+movie_id\+'(.+?)';`); re.MatchString(e.Text) {
			// var posterImage = '/moviepages/'+movie_id+'/images/main_b.jpg';
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 3 {
				info.ThumbURL = e.Request.AbsoluteURL(ss[1] + id + ss[2])
				info.CoverURL = info.ThumbURL /* use thumb as cover */
			}
		} else if re = regexp.MustCompile(`Movie\s*=\s*(\{.+?});`); re.MatchString(e.Text) {
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 2 {
				data := struct {
					SampleFlashURL  string `json:"sample_flash_url"`
					SampleMFlashURL string `json:"sample_m_flash_url"`
				}{}
				if json.Unmarshal([]byte(ss[1]), &data) == nil {
					for _, sample := range []string{
						data.SampleFlashURL, data.SampleMFlashURL,
					} {
						if sample != "" {
							info.PreviewVideoURL = e.Request.AbsoluteURL(sample)
							break
						}
					}
				}
			}
		}
	})

	// Preview Images
	c.OnXML(`//div[@class="gallery-ratio"]/a`, func(e *colly.XMLElement) {
		if href := e.Attr("href"); !strings.Contains(href, "member") {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(href))
		}
	})

	err = c.Visit(info.Homepage)
	return
}
