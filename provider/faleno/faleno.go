package faleno

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*FALENO)(nil)
	_ provider.MovieSearcher = (*FALENO)(nil)
)

const (
	Name     = "FALENO"
	Priority = 1000 - 4
)

const (
	baseURL   = "https://faleno.jp/top/"
	movieURL  = "https://faleno.jp/top/works/%s/"
	searchURL = "https://faleno.jp/top/?s=%s"
)

type FALENO struct {
	*scraper.Scraper
}

func New() *FALENO {
	return &FALENO{scraper.NewDefaultScraper(Name, baseURL, Priority, scraper.WithCookies(baseURL, []*http.Cookie{
		{Name: "modal", Value: "off"},
	}))}
}

func (fln *FALENO) NormalizeMovieID(id string) string {
	return strings.ToLower(id)
}

func (fln *FALENO) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return fln.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (fln *FALENO) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return fln.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (fln *FALENO) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := fln.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        id,
		Provider:      fln.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
	}

	c := fln.ClonedCollector()

	// Title
	c.OnXML(`//div[@class="bar02"]/h1/text()`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
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
			info.Actors = strings.Split(e.ChildText(`.//p`), "/")
		case "収録時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//p`))
		case "発売日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//p`))
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

func (fln *FALENO) NormalizeMovieKeyword(keyword string) string {
	if !regexp.MustCompile(`^(?i)f(s|c)dss-?\d{3}$`).MatchString(keyword) {
		return ""
	}
	return strings.ToLower(strings.ReplaceAll(keyword, "-", ""))
}

func (fln *FALENO) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := fln.ClonedCollector()
	c.ParseHTTPErrorResponse = true
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	c.OnXML(`//div[@class="box_kanren01"]//li`, func(e *colly.XMLElement) {
		cover := e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))

		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href"))
		id, _ := fln.ParseMovieIDFromURL(homepage)
		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      id,
			Title:       strings.SplitN(e.ChildText(`.//div[@class="text_name"]/a`), "\n", 2)[0],
			Provider:    fln.Name(),
			Homepage:    homepage,
			CoverURL:    cover,
			ReleaseDate: parser.ParseDate(strings.Fields(e.ChildText(`.//div[contains(text(), "発売開始")]`))[0]),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
