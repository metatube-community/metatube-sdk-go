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
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
	"golang.org/x/net/html"
)

var _ provider.MovieProvider = (*XxxAV)(nil)

const (
	Name     = "XXX-AV"
	Priority = 1000
)

const (
	baseURL  = "https://www.xxx-av.com/"
	movieURL = "https://www.xxx-av.com/mov/movie/%s/"
)

type XxxAV struct {
	*scraper.Scraper
}

func New() *XxxAV {
	return &XxxAV{
		Scraper: scraper.NewDefaultScraper(Name, Priority,
			scraper.WithCookies(baseURL, []*http.Cookie{
				{Name: "acc_accept_lang", Value: "japanese"},
			})),
	}
}

func (xav *XxxAV) NormalizeID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:XXX-AV-)?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (xav *XxxAV) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return xav.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (xav *XxxAV) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		ID:            path.Base(homepage.Path),
		Number:        fmt.Sprintf("XXX-AV-%s", path.Base(homepage.Path)),
		Provider:      xav.Name(),
		Homepage:      homepage.String(),
		Maker:         "トリプルエックス",
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
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
			info.Series = strings.TrimSpace(e.ChildText(`.//dd`))
		case "キーワード:":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//dd`),
				(*[]string)(&info.Tags))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
