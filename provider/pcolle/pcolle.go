package pcolle

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*Pcolle)(nil)

const (
	Name     = "Pcolle"
	Priority = 1000
)

const (
	baseURL  = "https://www.pcolle.com/"
	movieURL = "https://www.pcolle.com/product/detail/?product_id=%s"
)

type Pcolle struct {
	*scraper.Scraper
}

func New() *Pcolle {
	return &Pcolle{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		scraper.WithCookies(baseURL, []*http.Cookie{
			{Name: "AGE_CONF", Value: "1"},
		}),
	)}
}

func (pcl *Pcolle) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:PCOLLE[-_])?([a-z\d]{9,})$`).FindStringSubmatch(id); len(ss) == 2 {
		return strings.ToLower(ss[1])
	}
	return ""
}

func (pcl *Pcolle) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return pcl.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (pcl *Pcolle) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return pcl.NormalizeMovieID(homepage.Query().Get("product_id")), nil
}

func (pcl *Pcolle) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := pcl.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("PCOLLE-%s", id),
		Provider:      pcl.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := pcl.ClonedCollector()

	// Fields
	c.OnXML(`//table//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//th`) {
		case "販売会員:":
			info.Maker = e.ChildText(`.//td`)
		case "カテゴリー:":
			var texts []string
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//td`), &texts)
			if len(texts) > 0 {
				info.Label = texts[0]
			}
		case "商品名:":
			info.Title = e.ChildText(`.//td`)
		case "商品ID:":
			// use url product_id as ID.
			// info.ID = e.ChildText(`.//td`)
		case "販売開始日:":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td`))
		}
	})

	// Title (fallback)
	c.OnXML(`//div[@class="title-04"]`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//section[@class="item_description"]`, func(e *colly.XMLElement) {
		if summary := e.ChildText(`.//p[@class="fo-14"]`); summary != "" {
			// preferred summary.
			info.Summary = summary
		} else {
			info.Summary = e.Text
		}
	})

	// Thumb+Cover
	c.OnXML(`//div[@class="item-content"]//div[@class="part1"]/article`, func(e *colly.XMLElement) {
		info.ThumbURL = e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href"))
		info.CoverURL = info.ThumbURL
	})

	// Genres
	c.OnXML(`//section[@class="item_tags"]//ul//li`, func(e *colly.XMLElement) {
		info.Genres = append(info.Genres, strings.TrimSpace(e.Text))
	})

	// Preview Images
	c.OnXML(`//section[@class="item_images"]//ul//li`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages,
			e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href")))
	})

	// fallbacks
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" && len(info.PreviewImages) > 0 {
			// cover fallback.
			info.CoverURL = info.PreviewImages[0]
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
