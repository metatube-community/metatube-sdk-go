package jav321

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*JAV321)(nil)
	_ provider.MovieSearcher = (*JAV321)(nil)
)

const (
	Name     = "JAV321"
	Priority = 1000 - 6
)

const (
	baseURL   = "https://www.jav321.com/"
	movieURL  = "https://www.jav321.com/video/%s"
	searchURL = "https://www.jav321.com/search"
)

type JAV321 struct {
	*scraper.Scraper
}

func New() *JAV321 {
	return &JAV321{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (jav *JAV321) SetRequestTimeout(_ time.Duration) {
	jav.Scraper.SetRequestTimeout(10 * time.Second)
}

func (jav *JAV321) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return jav.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (jav *JAV321) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (jav *JAV321) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := jav.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      jav.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := jav.ClonedCollector()

	// Title
	c.OnXML(`/html/body/div[2]/div[1]/div[1]/div[1]/h3/text()`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Title (fallback)
	c.OnXML(`//div[@class='panel-heading']/h3/text()`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`/html/body/div[2]/div[1]/div[1]/div[2]/div[3]/div/text()`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Summary (fallback)
	c.OnXML(`//div[@class="panel-body"]/div[@class="row"]/div[@class="col-md-12"]/text()`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		if summary := strings.TrimSpace(e.Text); summary != "" {
			info.Summary = summary
		}
	})

	// Thumb
	c.OnXML(`//div[@class="panel-body"]/div[@class="row"]/div[@class="col-md-3"]/img`, func(e *colly.XMLElement) {
		if src := e.Attr("src"); src != "" {
			info.ThumbURL = e.Request.AbsoluteURL(src)
		}
	})

	// Cover+Images
	c.OnXML(`//div[@class="col-xs-12 col-md-12"]/p/a/img[@class="img-responsive"]`, func(e *colly.XMLElement) {
		if src := e.Attr("src"); src != "" {
			src = e.Request.AbsoluteURL(src)
			if info.CoverURL == "" {
				info.CoverURL = src // JAV321 use first image as cover.
			} else {
				info.PreviewImages = append(info.PreviewImages, src)
			}
		}
	})

	// Actors
	c.OnXML(`//div[@class="thumbnail"]/a[contains(@href,"/star/")]`, func(e *colly.XMLElement) {
		if actor := strings.TrimSpace(e.Text); actor != "" {
			info.Actors = append(info.Actors, e.Text)
		}
	})

	// Number
	c.OnXML(`//b[contains(text(),"品番")]/following-sibling::node()[1]`, func(e *colly.XMLElement) {
		info.Number = strings.ToUpper(strings.TrimSpace(
			strings.TrimLeft(e.DOM.(*html.Node).Data, ":")))
	})

	// ReleaseDate
	c.OnXML(`//b[contains(text(),"配信開始日")]/following-sibling::node()[1]`, func(e *colly.XMLElement) {
		info.ReleaseDate = parser.ParseDate(
			strings.TrimLeft(e.DOM.(*html.Node).Data, ":"))
	})

	// Runtime
	c.OnXML(`//b[contains(text(),"収録時間")]/following-sibling::node()[1]`, func(e *colly.XMLElement) {
		info.Runtime = parser.ParseRuntime(
			strings.TrimLeft(e.DOM.(*html.Node).Data, ":"))
	})

	// Series
	c.OnXML(`//b[contains(text(),"シリーズ")]/following-sibling::a[starts-with(@href,'/series')]`, func(e *colly.XMLElement) {
		info.Series = strings.TrimSpace(e.Text)
	})

	// Maker
	c.OnXML(`//b[contains(text(),"メーカー")]/following-sibling::a[starts-with(@href,"/company")]`, func(e *colly.XMLElement) {
		info.Maker = strings.TrimSpace(e.Text)
	})

	// Genres
	c.OnXML(`//b[contains(text(),"ジャンル")]/following-sibling::a[starts-with(@href,"/genre")]`, func(e *colly.XMLElement) {
		info.Genres = append(info.Genres, strings.TrimSpace(e.Text))
	})

	// Actors (fallback)
	c.OnXML(`//b[contains(text(),"出演者")]/following-sibling::a[starts-with(@href,"/star")]`, func(e *colly.XMLElement) {
		if len(info.Actors) > 0 {
			return
		}
		info.Actors = append(info.Actors, strings.TrimSpace(e.Text))
	})

	// Actors (fallback again)
	c.OnXML(`//b[contains(text(),"出演者")]/following-sibling::node()[1]`, func(e *colly.XMLElement) {
		if n := e.DOM.(*html.Node); n.Type == html.TextNode && len(info.Actors) == 0 {
			if actor := strings.TrimSpace(strings.TrimLeft(n.Data, ":")); actor != "" {
				info.Actors = append(info.Actors, actor)
			}
		}
	})

	// Score
	c.OnXML(`//b[contains(text(),"平均評価")]/following-sibling::img/@data-original`, func(e *colly.XMLElement) {
		if ss := regexp.MustCompile(`(\d+)\.gif`).FindStringSubmatch(e.Text); len(ss) == 2 {
			info.Score = parser.ParseScore(ss[1]) / 10
		}
	})

	// Score (fallback)
	c.OnXML(`//b[contains(text(),"平均評価")]/following-sibling::node()[1]`, func(e *colly.XMLElement) {
		if n := e.DOM.(*html.Node); n.Type == html.TextNode && info.Score == 0 {
			info.Score = parser.ParseScore(
				strings.TrimLeft(n.Data, ":"))
		}
	})

	// Preview Video
	c.OnXML(`//div[@class="panel-body"]//video/source/@src`, func(e *colly.XMLElement) {
		if src := strings.TrimSpace(e.Text); src != "" {
			src = strings.ReplaceAll(src, "awscc3001.r18.com", "cc3001.dmm.co.jp")
			src = strings.ReplaceAll(src, "cc3001.r18.com", "cc3001.dmm.co.jp")
			info.PreviewVideoURL = e.Request.AbsoluteURL(src)
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (jav *JAV321) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) && !regexp.MustCompile(`^(?i)([a-z]{1,4}\d{2,4}|heyzo[-_].+)$`).MatchString(keyword) {
		return "" // JavBus has no those special contents.
	}
	return strings.ToUpper(keyword)
}

func (jav *JAV321) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := jav.ClonedCollector()
	c.ParseHTTPErrorResponse = true
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	c.OnResponse(func(r *colly.Response) {
		var loc *url.URL
		if loc, err = url.Parse(r.Request.AbsoluteURL(r.Headers.Get("Location"))); err != nil {
			return
		}
		if strings.HasPrefix(loc.Path, "/video") {
			var info *model.MovieInfo
			if info, err = jav.GetMovieInfoByURL(loc.String()); err != nil {
				return
			}
			results = append(results, info.ToSearchResult())
		}
		//if strings.HasPrefix(loc.Path, "/snp") {
		//	// ignore
		//}
	})

	if postErr := c.Post(searchURL, map[string]string{"sn": keyword}); postErr != nil {
		err = postErr
	}
	return
}

func init() {
	provider.Register(Name, New)
}
