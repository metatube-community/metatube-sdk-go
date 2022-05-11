package tokyohot

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"golang.org/x/net/html"
)

var _ provider.MovieProvider = (*TokyoHot)(nil)

const (
	name     = "tokyo-hot"
	priority = 8
)

const (
	baseURL   = "https://my.tokyo-hot.com/"
	movieURL  = "https://my.tokyo-hot.com/product/%s/?lang=ja"
	searchURL = "https://my.tokyo-hot.com/product/?q=%s&x=0&y=0&lang=ja"
)

type TokyoHot struct {
	*provider.Scraper
}

func New() *TokyoHot {
	return &TokyoHot{
		Scraper: provider.NewScraper(name, priority, colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.IgnoreRobotsTxt(),
			colly.UserAgent(random.UserAgent()))),
	}
}

func (th *TokyoHot) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return th.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (th *TokyoHot) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	id := path.Base(homepage.Path)

	info = &model.MovieInfo{
		ID:            id,
		Provider:      th.Name(),
		Homepage:      homepage.String(),
		Maker:         "TOKYO-HOT",
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := th.Collector()

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
			switch dt {
			case "出演者":
				for _, actor := range e.ChildTexts(fmt.Sprintf(`.//dd[%d]/a`, i+1)) {
					if actor != "" && actor != "不明" {
						info.Actors = append(info.Actors, actor)
					}
				}
			case "プレイ内容":
				for _, tag := range e.ChildTexts(fmt.Sprintf(`.//dd[%d]/a`, i+1)) {
					if tag != "" {
						info.Tags = append(info.Tags, tag)
					}
				}
			case "シリーズ":
				info.Series = e.ChildText(fmt.Sprintf(`.//dd[%d]/a`, i+1))
			case "レーベル":
				info.Publisher = e.ChildText(fmt.Sprintf(`.//dd[%d]/a`, i+1))
			case "配信開始日":
				info.ReleaseDate = parser.ParseDate(e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1)))
			case "収録時間":
				info.Duration = parser.ParseDuration(e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1)))
			case "作品番号":
				info.Number = e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1))
			}
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (th *TokyoHot) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	{ // pre-handle keyword
		if !regexp.MustCompile(`^(?i)[a-z]*\d+`).MatchString(keyword) {
			return nil, provider.ErrInvalidKeyword
		}
		keyword = strings.ToLower(keyword)
	}

	c := th.Collector()

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
			Provider: th.Name(),
			Homepage: homepage.String(),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.RegisterMovieFactory(name, New)
}
