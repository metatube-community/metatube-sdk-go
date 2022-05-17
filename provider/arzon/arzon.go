package arzon

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"golang.org/x/net/html"
)

var (
	_ provider.MovieProvider = (*ARZON)(nil)
	_ provider.MovieSearcher = (*ARZON)(nil)
	_ provider.Fetcher       = (*ARZON)(nil)
)

const (
	Name     = "arzon"
	Priority = 1000 - 3 // ARZON as secondary provider has lower priority than official ones.
)

const (
	baseURL   = "https://www.arzon.jp/"
	movieURL  = "https://www.arzon.jp/item_%s.html"
	searchURL = "https://www.arzon.jp/itemlist.html?&q=%s&t=all&m=all&s=all&mkt=all&disp=30&sort=-udate"
)

// ARZON needs `Referer` header when request to view resources.
type ARZON struct {
	*provider.Scraper
}

func New() *ARZON {
	return &ARZON{
		Scraper: provider.NewScraper(Name, Priority, colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.IgnoreRobotsTxt(),
			colly.UserAgent(random.UserAgent()))),
	}
}

func (az *ARZON) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return az.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (az *ARZON) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		ID:            az.parseID(homepage.Path),
		Provider:      az.Name(),
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := az.Collector()

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
			info.Publisher = e.ChildText(`.//td[2]`)
		case "シリーズ：":
			info.Series = e.ChildText(`.//td[2]`)
		case "監督：":
			info.Director = e.ChildText(`.//td[2]`)
		case "発売日：":
			info.ReleaseDate = parser.ParseDate(strings.Fields(e.ChildText(`.//td[2]`))[0])
		case "収録時間：":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td[2]`))
		case "品番：":
			info.Number = strings.Fields(e.ChildText(`.//td[2]`))[0]
		case "タグ：":
			// info.Tags = e.ChildTexts(`.//td[2]`)
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (az *ARZON) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	{ // pre-handle keyword
		if number.IsUncensored(keyword) {
			return nil, provider.ErrInvalidKeyword
		}
		keyword = strings.ToUpper(keyword)
	}

	c := az.Collector()

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
		ids = append(ids, az.parseID(e.ChildAttr(`.//a`, "href")))
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

func (az *ARZON) Fetch(u string) (*http.Response, error) {
	return fetch.Fetch(u, fetch.WithReferer(baseURL))
}

func (az *ARZON) parseID(s string) string {
	if ss := regexp.MustCompile(`item_(\d+)\.html$`).FindStringSubmatch(s); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
