// Deprecated: This provider is no longer available.
package arzon

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*ARZON)(nil)
	_ provider.MovieSearcher = (*ARZON)(nil)
	_ provider.Fetcher       = (*ARZON)(nil)
)

const (
	Name     = "ARZON"
	Priority = 1000 - 3 // ARZON as secondary provider has lower priority than official ones.
)

const (
	baseURL   = "https://www.arzon.jp/"
	movieURL  = "https://www.arzon.jp/item_%s.html"
	searchURL = "https://www.arzon.jp/itemlist.html?&q=%s&t=all&m=all&s=all&mkt=all&disp=30&sort=-udate"
)

// ARZON needs `Referer` header when request to view resources.
type ARZON struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *ARZON {
	return &ARZON{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority),
	}
}

func (az *ARZON) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return az.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (az *ARZON) ParseMovieIDFromURL(rawURL string) (id string, err error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if ss := regexp.MustCompile(`item_(\d+)\.html$`).
		FindStringSubmatch(homepage.Path); len(ss) == 2 {
		id = ss[1]
	}
	return
}

func (az *ARZON) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := az.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      az.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := az.ClonedCollector()

	// Age check
	c.OnHTML(`#warn > table > tbody > tr > td.yes > a`, func(e *colly.HTMLElement) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			e.Response.Body = r.Body // Replace HTTP body
		})
		d.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Title
	c.OnXML(`//*[@id="detail_new"]//div[@class="detail_title_new2"]/table/tbody/tr/td[2]/h1`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//*[@id="detail_new"]/table/tbody/tr/td[1]/table/tbody/tr[2]/td/div`, func(e *colly.XMLElement) {
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

	// Cover+Thumb
	c.OnXML(`//*[@id="detail_new"]/table/tbody/tr/td[1]/table/tbody/tr[1]/td[1]/a/img`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("src"))
		info.ThumbURL = strings.ReplaceAll(info.CoverURL, "L.jpg", "S.jpg")
	})

	// Preview Images
	c.OnXML(`//*[@id="detail_new"]//div[@class="sample_img"]/img`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("src")))
	})

	// Score
	c.OnXML(`//*[@id="detail_new"]//div[@class="value"]//li[@class="review"]/img`, func(e *colly.XMLElement) {
		if ss := regexp.MustCompile(`star(\d+)\.gif`).FindStringSubmatch(e.Attr("src")); len(ss) == 2 {
			info.Score = parser.ParseScore(ss[1])
		}
	})

	// Fields
	c.OnXML(`//*[@id="detail_new"]/table/tbody/tr/td[1]/table/tbody/tr[3]/td/div/table/tbody/tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "AV女優：", "タレント：":
			info.Actors = e.ChildTexts(`.//td[2]/a`)
		case "AVメーカー：", "アニメメーカー：", "イメージメーカー：":
			info.Maker = e.ChildText(`.//td[2]`)
		case "AVレーベル：", "アニメレーベル：", "イメージレーベル：":
			info.Label = e.ChildText(`.//td[2]`)
		case "シリーズ：":
			info.Series = e.ChildText(`.//td[2]`)
		case "監督：":
			info.Director = e.ChildText(`.//td[2]`)
		case "発売日：":
			if fields := strings.Fields(e.ChildText(`.//td[2]`)); len(fields) > 0 {
				info.ReleaseDate = parser.ParseDate(fields[0])
			}
		case "収録時間：":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td[2]`))
		case "品番：":
			if fields := strings.Fields(e.ChildText(`.//td[2]`)); len(fields) > 0 {
				// Number can be empty occasionally.
				info.Number = fields[0]
			}
		case "タグ：":
			// info.Genres = e.ChildTexts(`.//td[2]`)
		}
	})

	c.OnXML(`//*[@id="allstock"]//strong`, func(e *colly.XMLElement) {
		if info.Number != "" {
			// Number is already fetched.
			return
		}
		if e.Text == "品番：" && e.DOM.(*html.Node).NextSibling != nil {
			info.Number = e.DOM.(*html.Node).NextSibling.Data
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (az *ARZON) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

func (az *ARZON) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := az.ClonedCollector()

	// Age check
	c.OnHTML(`#warn > table > tbody > tr > td.yes > a`, func(e *colly.HTMLElement) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			e.Response.Body = r.Body // Replace HTTP body
		})
		d.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	var ids []string
	c.OnXML(`//*[@id="item"]//dt`, func(e *colly.XMLElement) {
		id, _ := az.ParseMovieIDFromURL(e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href")))
		ids = append(ids, id)
	})

	c.OnScraped(func(r *colly.Response) {
		const limit = 3
		var (
			mu sync.Mutex
			wg sync.WaitGroup
		)
		for i := 0; i < len(ids) && i < limit; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if info, _ := az.GetMovieInfoByID(ids[i]); info != nil && info.Valid() {
					mu.Lock()
					results = append(results, info.ToSearchResult())
					mu.Unlock()
				}
			}(i)
		}
		wg.Wait()
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.Register(Name, New)
}
