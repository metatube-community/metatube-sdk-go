package prestige

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*PRESTIGE)(nil)
	_ provider.MovieSearcher = (*PRESTIGE)(nil)
)

const (
	Name     = "PRESTIGE"
	Priority = 1000 - 1
)

const (
	baseURL   = "https://www.prestige-av.com/"
	movieURL  = "https://www.prestige-av.com/goods/goods_detail.php?sku=%s"
	searchURL = "https://www.prestige-av.com/goods/goods_list.php?mode=free&mid=&word=%s&count=100&sort=near"
)

type PRESTIGE struct {
	*scraper.Scraper
}

func New() *PRESTIGE {
	return &PRESTIGE{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority,
			scraper.WithCookies(baseURL, []*http.Cookie{
				{Name: "coc", Value: "1"},
				{Name: "age_auth", Value: "1"},
			})),
	}
}

func (pst *PRESTIGE) NormalizeMovieID(id string) string {
	// PRESTIGE doesn't care about SKU cases, but we
	// use uppercase for better alignment.
	return strings.ToUpper(id)
}

func (pst *PRESTIGE) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return pst.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (pst *PRESTIGE) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return pst.NormalizeMovieID(homepage.Query().Get("sku")), nil
}

func (pst *PRESTIGE) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	id, err := pst.ParseMovieIDFromURL(u)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      pst.Name(),
		Homepage:      u,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := pst.ClonedCollector()

	// Title
	c.OnXML(`//div[@class="product_title_layout_01"]/h1`, func(e *colly.XMLElement) {
		for n := e.DOM.(*html.Node).FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				info.Title = strings.TrimSpace(n.Data)
			}
		}
		if info.Title == "" /* fallback */ {
			info.Title = trimTitle(e.Text)
		}
	})

	// Summary
	c.OnXML(`//div[@class="product_description_layout_01"]/p`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Thumb+Cover
	c.OnXML(`//div[@class="product_detail_layout_01"]//a[@class="sample_image"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
		info.ThumbURL = e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))
	})

	// Preview Video
	c.OnXML(`//*[@id="modal-main"]/video/source`, func(e *colly.XMLElement) {
		info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Preview Images
	c.OnXML(`//ul[@class="contents"]/li`, func(e *colly.XMLElement) {
		if href := e.ChildAttr(`.//a`, "href"); href != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(href))
		}
	})

	// Actor
	c.OnXML(`//dt[text()='出演：']/following-sibling::dd[1]`, func(e *colly.XMLElement) {
		var actors []string
		parser.ParseTexts(e.DOM.(*html.Node), &actors)
		for _, actor := range actors {
			for _, actor = range strings.Split(actor, "\u00a0" /* nbsp */) {
				// Remove redundant space from actor name.
				if actor = parser.ReplaceSpaceAll(actor); actor != "" {
					info.Actors = append(info.Actors, actor)
				}
			}
		}
	})

	// Fields
	c.OnXML(`//div[@class="product_detail_layout_01"]//dl[@class="spec_layout"]`, func(e *colly.XMLElement) {
		for i, dt := range e.ChildTexts(`.//dt`) {
			var (
				dd  = fmt.Sprintf(`.//dd[%d]`, i+1)
				dda = fmt.Sprintf(`.//dd[%d]/a`, i+1)
			)
			switch dt {
			case "収録時間：":
				info.Runtime = parser.ParseRuntime(e.ChildText(dd))
			case "発売日：":
				info.ReleaseDate = parser.ParseDate(e.ChildText(dda))
			case "メーカー名：":
				info.Maker = e.ChildText(dd)
			case "品番：":
				info.Number = strings.TrimSpace(e.ChildText(dd))
			case "ジャンル：":
				info.Genres = e.ChildTexts(dda)
			case "シリーズ：":
				info.Series = strings.TrimSpace(e.ChildText(dd))
			case "レーベル：":
				info.Label = strings.TrimSpace(e.ChildText(dd))
			}
		}
	})

	// Try to replace with high resolution pictures.
	c.OnScraped(func(_ *colly.Response) {
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				d := c.Clone()
				// existence check.
				switch i {
				case 0: // thumb
					d.OnScraped(func(r *colly.Response) {
						info.BigThumbURL = r.Request.URL.String()
					})
					d.Head(strings.ReplaceAll(info.ThumbURL, "/pf_p_", "/pf_"))
				case 1: // cover
					d.OnScraped(func(r *colly.Response) {
						info.BigCoverURL = r.Request.URL.String()
					})
					d.Head(strings.ReplaceAll(info.CoverURL, "/pb_e_", "/pb_"))
				}
			}(i)
		}
		wg.Wait()
	})

	err = c.Visit(info.Homepage)
	return
}

func (pst *PRESTIGE) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

func (pst *PRESTIGE) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := pst.ClonedCollector()

	c.OnXML(`//*[@id="body_goods"]/ul/li`, func(e *colly.XMLElement) {
		thumb := e.ChildAttr(`.//a/img`, "src")

		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href"))
		id, _ := pst.ParseMovieIDFromURL(homepage)

		var title string // colly.XMLElement takes all texts from elem, so we need to filter extra texts.
		for n := htmlquery.FindOne(e.DOM.(*html.Node), `.//a/span`).
			FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				// Normally, the last child is the title.
				title = strings.TrimSpace(n.Data)
			}
		}
		if title == "" /* fallback */ {
			title = trimTitle(e.ChildText(`.//a/span`))
		}

		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   id,
			Provider: pst.Name(),
			Title:    title,
			ThumbURL: e.Request.AbsoluteURL(imageSrc(thumb, true)),
			CoverURL: e.Request.AbsoluteURL(imageSrc(thumb, false)),
			Homepage: homepage,
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func trimTitle(s string) string {
	t := strings.Split(s, "\t")
	return strings.TrimSpace(t[len(t)-1])
}

func imageSrc(s string, thumb bool) string {
	if re := regexp.MustCompile(`(?i)/p[f|b]_[a-z\d]+?_`); re.MatchString(s) {
		if thumb {
			return re.ReplaceAllString(s, "/pf_p_")
		}
		return re.ReplaceAllString(s, "/pb_e_")
	}
	return s
}

func init() {
	provider.Register(Name, New)
}
