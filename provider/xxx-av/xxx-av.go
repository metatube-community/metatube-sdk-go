package xxx_av

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*XXXAV)(nil)

const (
	Name     = "XXX-AV"
	Priority = 1000
)

const (
	baseURL  = "https://www.xxx-av.com/"
	movieURL = "https://www.xxx-av.com/mov/movie/%s/"
)

type XXXAV struct {
	*scraper.Scraper
}

func New() *XXXAV {
	return &XXXAV{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority,
			scraper.WithCookies(baseURL, []*http.Cookie{
				{Name: "acc_accept_lang", Value: "japanese"},
			})),
	}
}

func (xav *XXXAV) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:xxx[-_]av[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (xav *XXXAV) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return xav.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (xav *XXXAV) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (xav *XXXAV) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := xav.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("XXX-AV-%s", id),
		Provider:      xav.Name(),
		Homepage:      rawURL,
		Maker:         "トリプルエックス",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := xav.ClonedCollector()

	// Title
	c.OnXML(`//div[@class="main_contents"]/h2`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//div[@class="main_contents"]//p[@class="mov_com"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Thumb+Cover
	c.OnXML(`//*[@id="streaming_player"]/img`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("src"))
		info.ThumbURL = info.CoverURL // same as cover
	})

	// Preview Images
	c.OnXML(`//div[@class="main_contents"]//div[@class="movie_sample_img "]/ul/li`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages,
			e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href")))
	})

	// Fields
	c.OnXML(`//div[@class="main_contents"]//dl[@class="info_dl clearfix"]`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//dt`) {
		case "公開日:":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//dd`))
		case "女優名:":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//dd`),
				(*[]string)(&info.Actors))
		case "再生時間:":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//dd`))
		case "カテゴリ名:":
			info.Label = strings.TrimSpace(e.ChildText(`.//dd`))
		case "キーワード:":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//dd`),
				(*[]string)(&info.Genres))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
