package tokyohot

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*TokyoHot)(nil)
	_ provider.MovieSearcher = (*TokyoHot)(nil)
)

const (
	Name     = "TOKYO-HOT"
	Priority = 1000 - 2
)

const (
	baseURL   = "https://my.tokyo-hot.com/"
	movieURL  = "https://my.tokyo-hot.com/product/%s/?lang=ja"
	searchURL = "https://my.tokyo-hot.com/product/?q=%s&x=0&y=0&lang=ja"
)

type TokyoHot struct {
	*scraper.Scraper
}

func New() *TokyoHot {
	return &TokyoHot{scraper.NewDefaultScraper(Name, baseURL, Priority)}
}

func (tht *TokyoHot) NormalizeID(id string) string {
	return strings.ToLower(id) // Tokyo-Hot uses lowercase ID.
}

func (tht *TokyoHot) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return tht.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (tht *TokyoHot) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (tht *TokyoHot) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := tht.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      tht.Name(),
		Homepage:      rawURL,
		Maker:         "TOKYO-HOT",
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := tht.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="main"]//div[@class="contents"]/h2`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//*[@id="main"]//div[@class="sentence"]`, func(e *colly.XMLElement) {
		var sentences []string
		for n := e.DOM.(*html.Node).FirstChild; n != nil; n = n.NextSibling {
			if n.Type != html.TextNode {
				continue
			}
			sentences = append(sentences, strings.TrimSpace(n.Data))
		}
		if len(sentences) > 0 {
			info.Summary = strings.TrimSpace(strings.Join(sentences, "\n"))
		}
	})

	// Image+Video
	c.OnXML(`//div[@class="flowplayer"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.ChildAttr(`.//video`, "poster")
		info.ThumbURL = info.CoverURL // same as cover
		info.PreviewVideoURL = e.ChildAttr(`.//source`, "src")
	})

	// Preview Images
	c.OnXML(`//a[@rel="cap"]`, func(e *colly.XMLElement) {
		if href := e.Attr("href"); href != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(href))
		}
	})

	// Fields
	c.OnXML(`//*[@id="main"]//div[@class="infowrapper"]/dl`, func(e *colly.XMLElement) {
		for i, dt := range e.ChildTexts(`.//dt`) {
			var (
				dd  = fmt.Sprintf(`.//dd[%d]`, i+1)
				dda = fmt.Sprintf(`.//dd[%d]/a`, i+1)
			)
			switch dt {
			case "出演者":
				var actors []string
				parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), dd), &actors)
				for _, actor := range actors {
					if actor != "不明" {
						info.Actors = append(info.Actors, actor)
					}
				}
			case "プレイ内容":
				parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), dd),
					(*[]string)(&info.Tags))
			case "シリーズ":
				info.Series = e.ChildText(dda)
			case "レーベル":
				info.Publisher = e.ChildText(dda)
			case "配信開始日":
				info.ReleaseDate = parser.ParseDate(e.ChildText(dd))
			case "収録時間":
				info.Runtime = parser.ParseRuntime(e.ChildText(dd))
			case "作品番号":
				info.Number = e.ChildText(dd)
			}
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (tht *TokyoHot) TidyKeyword(keyword string) string {
	if regexp.MustCompile(`^(?i)[a-z_]*\d+`).MatchString(keyword) {
		return strings.ToLower(keyword)
	}
	return ""
}

func (tht *TokyoHot) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := tht.ClonedCollector()

	c.OnXML(`//*[@id="main"]/ul/li`, func(e *colly.XMLElement) {
		img := e.ChildAttr(`.//a/img`, "src")
		href := e.ChildAttr(`.//a`, "href")
		homepage, _ := url.Parse(e.Request.AbsoluteURL(href))
		results = append(results, &model.MovieSearchResult{
			ID:       path.Base(homepage.Path),
			Number:   path.Base(homepage.Path),
			Title:    e.ChildText(`.//div[@class="title"]`),
			ThumbURL: e.Request.AbsoluteURL(img),
			CoverURL: e.Request.AbsoluteURL(img),
			Provider: tht.Name(),
			Homepage: homepage.String(),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
