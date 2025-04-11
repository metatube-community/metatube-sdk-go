package gcolle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*Gcolle)(nil)

const (
	Name     = "Gcolle"
	Priority = 1000
)

const (
	baseURL  = "https://gcolle.net/"
	movieURL = "https://gcolle.net/product_info.php/products_id/%s"
	scoreURL = "https://rating.gcolle.net/ratings/products/%s.js"
)

type Gcolle struct {
	*scraper.Scraper
}

func New() *Gcolle {
	return &Gcolle{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		scraper.WithDetectCharset(),
	)}
}

func (gcl *Gcolle) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:GCOLLE[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (gcl *Gcolle) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return gcl.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (gcl *Gcolle) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (gcl *Gcolle) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := gcl.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("GCOLLE-%s", id),
		Provider:      gcl.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := gcl.ClonedCollector()

	// Age check
	c.OnHTML(`#main_content > table:nth-child(5) > tbody > tr > td:nth-child(2) > table > tbody > tr > td > h4 > a:nth-child(2)`, func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if !strings.Contains(href, "age_check") {
			return
		}
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			e.Response.Body = r.Body // Replace HTTP body.
		})
		d.Visit(e.Request.AbsoluteURL(href))
	})

	// Title
	c.OnXML(`//*[@id="cart_quantity"]/table/tbody/tr[1]/td/h1`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//*[@id="cart_quantity"]/table/tbody/tr[3]/td/p`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Genres
	c.OnXML(`//*[@id="cart_quantity"]/table/tbody/tr[4]/td/a`, func(e *colly.XMLElement) {
		info.Genres = append(info.Genres, strings.TrimSpace(e.Text))
	})

	// Thumb+Cover
	c.OnXML(`//*[@id="cart_quantity"]/table/tbody/tr[3]/td/table/tbody/tr/td/a`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
		info.ThumbURL = e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))
	})

	// Preview Images
	c.OnXML(`//*[@id="cart_quantity"]/table/tbody/tr[3]/td/div/img`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages,
			e.Request.AbsoluteURL(e.Attr("src")))
	})

	// Preview Images (extra?)
	c.OnXML(`//*[@id="cart_quantity"]/table/tbody/tr[3]/td/div/a/img`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages,
			e.Request.AbsoluteURL(e.Attr("src")))
	})

	// Fields
	c.OnXML(`//table[@class="filesetumei"]//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "商品番号":
			// should use id from url.
			// info.ID = e.ChildText(`.//td[2]`)
		case "商品登録日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
		}
	})

	// Maker
	c.OnXML(`//table[@class="contentBoxContentsManufactureInfo"]//td`, func(e *colly.XMLElement) {
		if info.Maker != "" {
			return
		}
		if strings.Contains(e.Text, "アップロード会員名") {
			info.Maker = e.ChildText(`.//b`)
		}
	})

	// Score
	c.OnScraped(func(_ *colly.Response) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			data := struct {
				Rating float64 `json:"rating"`
			}{}
			if json.Unmarshal(r.Body, &data) == nil {
				info.Score = data.Rating
			}
		})
		d.Visit(fmt.Sprintf(scoreURL, id))
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
