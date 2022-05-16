package prestige

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"golang.org/x/net/html"
)

var (
	_ provider.MovieProvider = (*PRESTIGE)(nil)
	_ provider.MovieSearcher = (*PRESTIGE)(nil)
)

const (
	Name     = "prestige"
	Priority = 10
)

const (
	baseURL   = "https://www.prestige-av.com/"
	movieURL  = "https://www.prestige-av.com/goods/goods_detail.php?sku=%s"
	searchURL = "https://www.prestige-av.com/goods/goods_list.php?mode=free&mid=&word=%s&count=100&sort=near"
)

type PRESTIGE struct {
	*provider.Scraper
}

func New() *PRESTIGE {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.IgnoreRobotsTxt(),
		colly.UserAgent(random.UserAgent()))
	c.SetCookies(baseURL, []*http.Cookie{
		{Name: "coc", Value: "1"},
		{Name: "age_auth", Value: "1"},
	})
	return &PRESTIGE{provider.NewScraper(Name, Priority, c)}
}

func (pst *PRESTIGE) NormalizeID(id string) string {
	// PRESTIGE doesn't case about SKU cases, but we
	// use uppercase for better alignment.
	return strings.ToUpper(id)
}

func (pst *PRESTIGE) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return pst.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (pst *PRESTIGE) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		Provider:      pst.Name(),
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	if info.ID = strings.ToUpper(homepage.Query().Get("sku")); info.ID == "" {
		return nil, provider.ErrInvalidID
	}

	c := pst.Collector()

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
		var getActor func(*html.Node)
		getActor = func(n *html.Node) {
			if n.Type == html.TextNode {
				if actor := replaceSpaceAll(n.Data); actor != "" {
					info.Actors = append(info.Actors, actor)
				}
			}
			for n := n.FirstChild; n != nil; n = n.NextSibling {
				getActor(n)
			}
		}
		getActor(e.DOM.(*html.Node))
	})

	// Fields
	c.OnXML(`//div[@class="product_detail_layout_01"]//dl[@class="spec_layout"]`, func(e *colly.XMLElement) {
		for i, dt := range e.ChildTexts(`.//dt`) {
			switch dt {
			case "収録時間：":
				info.Runtime = parser.ParseRuntime(e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1)))
			case "発売日：":
				info.ReleaseDate = parser.ParseDate(e.ChildText(fmt.Sprintf(`.//dd[%d]/a`, i+1)))
			case "メーカー名：":
				info.Maker = e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1))
			case "品番：":
				info.Number = strings.TrimSpace(e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1)))
			case "ジャンル：":
				info.Tags = e.ChildTexts(fmt.Sprintf(`.//dd[%d]/a`, i+1))
			case "シリーズ：":
				info.Series = strings.TrimSpace(e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1)))
			case "レーベル：":
				info.Publisher = strings.TrimSpace(e.ChildText(fmt.Sprintf(`.//dd[%d]`, i+1)))
			}
		}
	})

	// Try to replace with high resolution pictures.
	c.OnScraped(func(_ *colly.Response) {
		var wg sync.WaitGroup
		for _, p := range []*string{
			&info.ThumbURL, &info.CoverURL,
		} {
			wg.Add(1)
			go func(p *string) {
				defer wg.Done()
				d := c.Clone()
				d.OnScraped(func(r *colly.Response) {
					*p = r.Request.URL.String()
				})
				// see if image exists.
				switch {
				case strings.Contains(*p, "/pf_p_"): // thumb
					_ = d.Head(strings.ReplaceAll(*p, "/pf_p_", "/pf_"))
				case strings.Contains(*p, "/pb_e_"): // cover
					_ = d.Head(strings.ReplaceAll(*p, "/pb_e_", "/pb_"))
				}
			}(p)
		}
		wg.Wait()
	})

	err = c.Visit(info.Homepage)
	return
}

func (pst *PRESTIGE) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	{ // pre-handle keyword
		if number.IsUncensored(keyword) {
			return nil, provider.ErrInvalidKeyword
		}
		keyword = strings.ToUpper(keyword)
	}

	c := pst.Collector()

	c.OnXML(`//*[@id="body_goods"]/ul/li`, func(e *colly.XMLElement) {
		href := e.ChildAttr(`.//a`, "href")
		thumb := e.ChildAttr(`.//a/img`, "src")
		var id string
		if ss := regexp.MustCompile(`(?i)sku=([a-z\d-]+)`).FindStringSubmatch(href); len(ss) == 2 {
			id = strings.ToUpper(ss[1])
		} else {
			return // ignore this one.
		}

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
			Homepage: e.Request.AbsoluteURL(href),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func trimTitle(s string) string {
	t := strings.Split(s, "\t")
	return strings.TrimSpace(t[len(t)-1])
}

func replaceSpaceAll(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, c := range s {
		if !unicode.IsSpace(c) {
			b.WriteRune(c)
		}
	}
	return b.String()
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
	provider.RegisterMovieFactory(Name, New)
}
