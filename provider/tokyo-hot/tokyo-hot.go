package tokyohot

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/collection/sets"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
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
	return &TokyoHot{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		// Temporary workaround for unknown CA issue.
		scraper.WithTransport(&http.Transport{
			Proxy:                 http.DefaultTransport.(*http.Transport).Proxy,
			DialContext:           http.DefaultTransport.(*http.Transport).DialContext,
			ForceAttemptHTTP2:     http.DefaultTransport.(*http.Transport).ForceAttemptHTTP2,
			MaxIdleConns:          http.DefaultTransport.(*http.Transport).MaxIdleConns,
			IdleConnTimeout:       http.DefaultTransport.(*http.Transport).IdleConnTimeout,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		}),
	)}
}

func (tht *TokyoHot) NormalizeMovieID(id string) string {
	return strings.ToLower(id) /* Tokyo-Hot uses lowercase ID */
}

func (tht *TokyoHot) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return tht.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (tht *TokyoHot) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return tht.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (tht *TokyoHot) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := tht.ParseMovieIDFromURL(rawURL)
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
		Genres:        []string{},
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

	// Thumb+Cover
	c.OnXML(`//li[@class="package"]`, func(e *colly.XMLElement) {
		for _, href := range e.ChildAttrs(`.//a`, "href") {
			href = e.Request.AbsoluteURL(href)
			if info.CoverURL == "" &&
				(strings.HasSuffix(href, "L.jpg") ||
					strings.Contains(href, "jacket")) {
				info.CoverURL = href
			} else if info.ThumbURL == "" &&
				(strings.HasSuffix(href, "v.jpg") ||
					strings.HasSuffix(href, "vb.jpg") ||
					strings.Contains(href, "package")) {
				info.ThumbURL = href
			}
		}
	})

	// Cover (fallback) + Video
	c.OnXML(`//div[@class="flowplayer"]`, func(e *colly.XMLElement) {
		if poster := e.ChildAttr(`.//video`, "poster"); info.CoverURL == "" && poster != "" {
			info.CoverURL = e.Request.AbsoluteURL(poster)
		}
		if src := e.ChildAttr(`.//source`, "src"); src != "" {
			info.PreviewVideoURL = e.Request.AbsoluteURL(src)
		}
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
				var genres []string
				parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), dd), &genres)
				info.Genres = append(info.Genres, genres...)
			case "タグ":
				// Additional Genres info
				var genres []string
				parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), dd), &genres)
				info.Genres = append(info.Genres, genres...)
			case "シリーズ":
				info.Series = e.ChildText(dda)
			case "レーベル":
				info.Label = e.ChildText(dda)
			case "配信開始日":
				info.ReleaseDate = parser.ParseDate(e.ChildText(dd))
			case "収録時間":
				info.Runtime = parser.ParseRuntime(e.ChildText(dd))
			case "作品番号":
				info.Number = e.ChildText(dd)
			}
		}
	})

	// Deduplicate Genres
	c.OnScraped(func(_ *colly.Response) {
		genres := sets.NewOrderedSet[string]()
		for _, genre := range info.Genres {
			genres.Add(genre)
		}
		info.Genres = genres.AsSlice()
	})

	// Fallbacks
	c.OnScraped(func(_ *colly.Response) {
		switch {
		case info.ThumbURL == "" && info.CoverURL != "":
			info.ThumbURL = info.CoverURL // use cover as thumb.
		case info.CoverURL == "" && info.ThumbURL != "":
			info.CoverURL = info.ThumbURL // vice versa.
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (tht *TokyoHot) NormalizeMovieKeyword(keyword string) string {
	if regexp.MustCompile(`^(?i)[a-z_]*\d+$`).MatchString(keyword) {
		return strings.ToLower(keyword)
	}
	return ""
}

func (tht *TokyoHot) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := tht.ClonedCollector()

	c.OnXML(`//*[@id="main"]/ul/li`, func(e *colly.XMLElement) {
		img := e.Request.AbsoluteURL(e.ChildAttr(`.//a/img`, "src"))
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href"))
		id, _ := tht.ParseMovieIDFromURL(homepage)

		// id is not always the number.
		var number string
		if ss := regexp.MustCompile(`\(作品番号: ([a-z\d-_]+)\)`).
			FindStringSubmatch(e.Text); len(ss) == 2 {
			number = ss[1]
		}
		{ // number fallbacks
			if number == "" {
				number = e.ChildAttr(`.//a/img`, "alt")
			}
			if number == "" {
				number = e.ChildAttr(`.//a/img`, "title")
			}
			if number == "" {
				number = id
			}
		}
		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   number,
			Title:    e.ChildText(`.//div[@class="title"]`),
			ThumbURL: img,
			CoverURL: img,
			Provider: tht.Name(),
			Homepage: homepage,
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.Register(Name, New)
}
