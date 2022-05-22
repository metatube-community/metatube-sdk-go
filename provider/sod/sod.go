package sod

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*SOD)(nil)
	_ provider.MovieSearcher = (*SOD)(nil)
	_ provider.Fetcher       = (*SOD)(nil)
)

const (
	Name     = "SOD"
	Priority = 1000
)

const (
	baseURL   = "https://ec.sod.co.jp/prime/"
	movieURL  = "https://ec.sod.co.jp/prime/videos/?id=%s"
	searchURL = "https://ec.sod.co.jp/prime/videos/genre/?search_type=1&sodsearch=%s"
	onTimeURL = "https://ec.sod.co.jp/prime/_ontime.php"
)

// SOD needs `Referer` header when request to view images and videos.
type SOD struct {
	*scraper.Scraper
}

func New() *SOD {
	return &SOD{scraper.NewDefaultScraper(Name, Priority)}
}

func (sod *SOD) NormalizeID(id string) string {
	return strings.ToUpper(id) // SOD requires uppercase ID.
}

func (sod *SOD) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return sod.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (sod *SOD) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		Provider:      sod.Name(),
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	// ID+Number
	if ss := regexp.MustCompile(`id=(.+?)$`).FindStringSubmatch(info.Homepage); len(ss) == 2 {
		info.ID = strings.ToUpper(ss[1])
		info.Number = info.ID
	}

	c := sod.ClonedCollector()
	composedMovieURL := fmt.Sprintf(movieURL, url.QueryEscape(info.ID))

	// Age check
	c.OnHTML(`#modal > div.pkg_age > div.enter > a`, func(e *colly.HTMLElement) {
		d := c.Clone()
		d.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Referer", composedMovieURL)
		})
		d.OnResponse(func(r *colly.Response) {
			e.Response.Body = r.Body // Replace HTTP body
		})
		d.Visit(e.Request.AbsoluteURL(e.Attr("href"))) // onTime
	})

	// Fields
	c.OnXML(`//*[@id="v_introduction"]/tbody/tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "品番":
			info.Number = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "発売年月日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
		case "シリーズ名":
			info.Series = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "出演者":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//td[2]`),
				(*[]string)(&info.Actors))
		case "再生時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td[2]`))
		case "監督":
			info.Director = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "メーカー":
			info.Maker = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "レーベル":
			info.Publisher = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "ジャンル":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//td[2]`),
				(*[]string)(&info.Tags))
		}
	})

	// Title
	c.OnXML(`//p[@class="product_title"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//div[@class="videos_textli"]/article`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Preview Video
	c.OnXML(`//div[@class="videos_textli"]//div[@class="videos_sampb"]/a`, func(e *colly.XMLElement) {
		d := c.Clone()
		d.OnXML(`//*[@id="moviebox"]/video/source`, func(e *colly.XMLElement) {
			info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("src"))
		})
		d.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Thumb+Cover
	c.OnXML(`//*[@id="videos_toptable"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.ChildAttr(`.//div[@class="videos_samimg"]/a[1]`, "href")
		info.ThumbURL = e.ChildAttr(`.//div[@class="videos_samimg"]/a[1]/img`, "src")
	})

	// Preview Images
	c.OnXML(`//*[@id="videos_samsbox"]/a`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Score
	c.OnXML(`//*[@id="review_body"]//div[@class="imagestar"]/i`, func(e *colly.XMLElement) {
		info.Score = parser.ParseScore(e.Text)
	})

	err = c.Visit(composedMovieURL)
	return
}

func (sod *SOD) TidyKeyword(keyword string) string {
	if !number.IsUncensored(keyword) {
		return strings.ToUpper(keyword)
	}
	return ""
}

func (sod *SOD) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := sod.ClonedCollector()
	composedSearchURL := fmt.Sprintf(searchURL, url.QueryEscape(keyword))

	// Age check
	c.OnHTML(`#modal > div.pkg_age > div.enter > a`, func(e *colly.HTMLElement) {
		d := c.Clone()
		d.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Referer", composedSearchURL)
		})
		d.OnResponse(func(r *colly.Response) {
			e.Response.Body = r.Body // Replace HTTP body
		})
		d.Visit(e.Request.AbsoluteURL(e.Attr("href"))) // onTime
	})

	c.OnXML(`//*[@id="videos_s_mainbox"]`, func(e *colly.XMLElement) {
		searchResult := &model.MovieSearchResult{
			Title:       e.ChildText(`.//div[@class="videis_s_txt"]/h2/a`),
			Provider:    sod.Name(),
			Homepage:    e.Request.AbsoluteURL(e.ChildAttr(`.//div[@class="videis_s_img"]/a`, "href")),
			ReleaseDate: parser.ParseDate(e.ChildText(`.//div[@class="videis_s_star"]/p`)),
		}

		// ID+Number
		if ss := regexp.MustCompile(`id=(.+?)$`).FindStringSubmatch(searchResult.Homepage); len(ss) == 2 {
			searchResult.ID = ss[1]
			searchResult.Number = searchResult.ID
		}

		// Thumb+Cover
		if thumb := e.ChildAttr(`.//div[@class="videis_s_img"]/a/img`, "src"); thumb != "" {
			searchResult.ThumbURL = e.Request.AbsoluteURL(thumb)
			searchResult.CoverURL = strings.ReplaceAll(searchResult.ThumbURL, "_m.jpg", "_l.jpg")
		}

		results = append(results, searchResult)
	})

	err = c.Visit(composedSearchURL)
	return
}

func (sod *SOD) Fetch(u string) (*http.Response, error) {
	return fetch.Fetch(u, fetch.WithReferer(baseURL))
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
