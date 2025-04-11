package sod

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*SOD)(nil)
	_ provider.MovieSearcher = (*SOD)(nil)
	_ provider.Fetcher       = (*SOD)(nil)
)

const (
	Name     = "SOD"
	Priority = 1000 - 1
)

const (
	baseURL   = "https://ec.sod.co.jp/prime/"
	movieURL  = "https://ec.sod.co.jp/prime/videos/?id=%s"
	searchURL = "https://ec.sod.co.jp/prime/videos/genre/?search_type=1&sodsearch=%s"
	onTimeURL = "https://ec.sod.co.jp/prime/_ontime.php"
)

var ErrImageNotAvailable = errors.New("image not available")

// SOD needs `Referer` header when request to view images and videos.
type SOD struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *SOD {
	return &SOD{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese),
	}
}

func (sod *SOD) NormalizeMovieID(id string) string {
	return strings.ToUpper(id) /* SOD requires uppercase ID */
}

func (sod *SOD) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return sod.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (sod *SOD) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return sod.NormalizeMovieID(homepage.Query().Get("id")), nil
}

func (sod *SOD) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := sod.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        id,
		Provider:      sod.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
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
			info.Label = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "ジャンル":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//td[2]`),
				(*[]string)(&info.Genres))
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

	defer func() {
		// Validate cover image
		if err == nil && !isValidImageURL(info.CoverURL) {
			err = ErrImageNotAvailable
		}
	}()

	err = c.Visit(composedMovieURL)
	return
}

func (sod *SOD) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
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
		thumb := e.Request.AbsoluteURL(e.ChildAttr(`.//div[@class="videis_s_img"]/a/img`, "src"))
		if !isValidImageURL(thumb) {
			return
		}
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//div[@class="videis_s_img"]/a`, "href"))
		id, _ := sod.ParseMovieIDFromURL(homepage)
		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      id,
			Title:       e.ChildText(`.//div[@class="videis_s_txt"]/h2/a`),
			Provider:    sod.Name(),
			Homepage:    homepage,
			ThumbURL:    thumb,
			CoverURL:    strings.ReplaceAll(thumb, "_m.jpg", "_l.jpg"),
			ReleaseDate: parser.ParseDate(e.ChildText(`.//div[@class="videis_s_star"]/p`)),
		})
	})

	err = c.Visit(composedSearchURL)
	return
}

func isValidImageURL(s string) bool {
	return !regexp.MustCompile(`/thumbnail/now_\w+\.jpg`).MatchString(s)
}

func init() {
	provider.Register(Name, New)
}
