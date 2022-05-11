package sod

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
)

var (
	_ provider.MovieProvider = (*SOD)(nil)
	_ provider.MovieSearcher = (*SOD)(nil)
	_ provider.Downloader    = (*SOD)(nil)
)

const (
	name     = "sod"
	priority = 10
)

const (
	baseURL   = "https://ec.sod.co.jp/prime/"
	movieURL  = "https://ec.sod.co.jp/prime/videos/?id=%s"
	searchURL = "https://ec.sod.co.jp/prime/videos/genre/?search_type=1&sodsearch=%s"
	onTimeURL = "https://ec.sod.co.jp/prime/_ontime.php"
)

// SOD needs `Referer` header when request to view images and videos.
type SOD struct {
	*provider.Scraper
}

func New() *SOD {
	return &SOD{
		Scraper: provider.NewScraper(name, priority, colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.IgnoreRobotsTxt(),
			colly.UserAgent(random.UserAgent()))),
	}
}

func (sod *SOD) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	id = strings.ToUpper(id) // SOD requires uppercase ID
	return sod.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (sod *SOD) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		Provider:      name,
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	// ID+Number
	if ss := regexp.MustCompile(`id=(.+?)$`).FindStringSubmatch(info.Homepage); len(ss) == 2 {
		info.ID = ss[1]
		info.Number = info.ID
	}

	c := sod.Collector()
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
			info.ID = e.ChildText(`.//td[2]`)
		case "発売年月日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
		case "シリーズ名":
			info.Series = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "出演者":
			info.Actors = e.ChildTexts(`.//td[2]/a`)
		case "再生時間":
			info.Duration = parser.ParseDuration(e.ChildText(`.//td[2]`))
		case "監督":
			info.Director = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "メーカー":
			info.Maker = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "レーベル":
			info.Publisher = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "ジャンル":
			info.Tags = e.ChildTexts(`.//td[2]`)
			if tags := e.ChildTexts(`.//td[2]/a`); len(tags) > 0 {
				info.Tags = tags
			}
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

	// Summary
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

func (sod *SOD) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	{ // pre-handle keyword
		if number.IsUncensored(keyword) {
			return nil, provider.ErrInvalidKeyword
		}
		keyword = strings.ToUpper(keyword) // SOD prefers uppercase
	}

	c := sod.Collector()
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
			Provider:    name,
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

func (sod *SOD) Download(u string) (_ io.ReadCloser, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	if req, err = http.NewRequest(http.MethodGet, u, nil); err != nil {
		return
	}
	// SOD needs referer header to view image/video
	req.Header.Set("Referer", baseURL)
	// make request
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	return resp.Body, nil
}

func init() {
	provider.RegisterMovieFactory(name, New)
}
