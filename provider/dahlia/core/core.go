package core

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
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
	BaseURL   string
	MovieURL  string
	SearchURL string

	// Values
	DefaultPriority float64
	DefaultName     string
}

func (core *Core) Init() *Core {
	core.Scraper = scraper.NewDefaultScraper(
		core.DefaultName,
		core.BaseURL,
		core.DefaultPriority,
		language.Japanese,
		scraper.WithCookies(core.BaseURL, []*http.Cookie{
			{Name: "modal", Value: "off"},
		}))
	return core
}

func (core *Core) NormalizeMovieID(id string) string { return strings.ToLower(id) }

func (core *Core) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return core.GetMovieInfoByURL(fmt.Sprintf(core.MovieURL, id))
}

func (core *Core) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return core.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (core *Core) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := core.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        parser.ParseIDToNumber(id),
		Provider:      core.Name(),
		Homepage:      rawURL,
		Maker:         core.Name(),
		Label:         core.Name(),
		Actors:        []string{},
		Genres:        []string{},
		PreviewImages: []string{},
	}

	c := core.ClonedCollector()

	// Title
	c.OnXML(`//div[@class="bar02_works"]/h1/text()`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Title (fallback)
	c.OnXML(`//div[@class="bar02"]/h1/text()`, func(e *colly.XMLElement) {
		if info.Title == "" {
			info.Title = strings.TrimSpace(e.Text)
		}
	})

	// Cover
	c.OnXML(`//div[@class="box_works01_img"]/a/img`, func(e *colly.XMLElement) {
		info.CoverURL = strings.Split(e.Request.AbsoluteURL(e.Attr("src")), "?")[0]
	})

	// Preview Video
	c.OnXML(`//div[@class="box_works01_img"]/a`, func(e *colly.XMLElement) {
		info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Previews
	c.OnXML(`//div[@class="box_works01_ga"]//li/a`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Summary
	c.OnXML(`//div[@class="box_works01_text"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		if summary := strings.TrimSpace(e.Text); summary != "" {
			info.Summary = strings.ReplaceAll(summary, "\n", "<br />")
		}
	})

	// Fields
	c.OnXML(`//div[contains(@class, "box_works01_list")]/ul/*[child::span or (@class="view_timer" and not(contains(@style,'display: none')))]//span/parent::*`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span`) {
		case "出演女優":
			if actors := e.ChildText(`.//p`); actors != "" {
				info.Actors = strings.Split(actors, "/")
			}
		case "収録時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//p`))
		case "配信開始日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//p`))
		case "発売日":
			if time.Time(info.ReleaseDate).IsZero() {
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//p`))
			}
		}
	})

	// ReleaseDate (fallback)
	c.OnXML(`//div[contains(@class, "box_works01_list")]/ul/div[@class="view_timer" and not(contains(@style,'display: none'))]/li/span[text()="配信開始日"]/following-sibling::p`, func(e *colly.XMLElement) {
		if time.Time(info.ReleaseDate).IsZero() {
			info.ReleaseDate = parser.ParseDate(e.Text)
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (core *Core) NormalizeMovieKeyword(keyword string) string {
	return strings.ToLower(strings.ReplaceAll(keyword, "-", ""))
}

func (core *Core) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := core.ClonedCollector()
	c.ParseHTTPErrorResponse = true
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	// Set max search timeout.
	// c.SetRequestTimeout(8 * time.Second)

	c.OnXML(`//div[@class="box_kanren01"]//li`, func(e *colly.XMLElement) {
		cover := e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))

		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href"))
		id, _ := core.ParseMovieIDFromURL(homepage)
		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      parser.ParseIDToNumber(id),
			Title:       strings.SplitN(e.ChildText(`.//div[@class="text_name"]/a`), "\n", 2)[0],
			Provider:    core.Name(),
			Homepage:    homepage,
			CoverURL:    cover,
			ReleaseDate: parser.ParseDate(strings.Fields(e.ChildText(`.//div[contains(text(), "発売開始")]`))[0]),
		})
	})

	err = c.Visit(fmt.Sprintf(core.SearchURL, url.QueryEscape(keyword)))
	return
}
