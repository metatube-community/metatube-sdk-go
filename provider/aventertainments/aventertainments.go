package aventertainments

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*AVE)(nil)
	_ provider.MovieSearcher = (*AVE)(nil)
)

const (
	Name     = "AVE"
	Priority = 1000 - 2
)

const (
	baseURL   = "https://www.aventertainments.com/"
	movieURL  = "https://www.aventertainments.com/product_lists.aspx?product_id=%s&languageID=2&dept_id=29"
	searchURL = "https://www.aventertainments.com/search_Products.aspx?languageID=2&dept_id=29&keyword=%s&searchby=keyword"
)

type AVE struct {
	*scraper.Scraper
}

func New() *AVE {
	return &AVE{scraper.NewDefaultScraper(Name, baseURL, Priority)}
}

func (ave *AVE) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return ave.GetMovieInfoByURL(fmt.Sprintf(movieURL, url.QueryEscape(id)))
}

func (ave *AVE) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return homepage.Query().Get("product_id"), nil
}

func (ave *AVE) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := ave.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      ave.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := ave.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="MyBody"]//div[@class="section-title"]/h3`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//*[@id="MyBody"]//div[@class="product-description mt-20"]`, func(e *colly.XMLElement) {
		for n := e.DOM.(*html.Node).FirstChild; n != nil; n = n.NextSibling {
			if n.Type == html.TextNode {
				info.Summary = strings.TrimSpace(n.Data)
				return
			}
		}
		// fallback
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Thumb+Cover
	c.OnXML(`//*[@id="PlayerCover"]/img`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("src"))
		info.ThumbURL = strings.ReplaceAll(info.CoverURL, "bigcover", "jacket_images")
	})

	// Thumb (fallback)
	c.OnXML(`//link[@rel="image_src"]`, func(e *colly.XMLElement) {
		if info.ThumbURL != "" {
			return
		}
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Cover (fallback)
	c.OnXML(`//span[@class="grid-gallery"]/a`, func(e *colly.XMLElement) {
		if info.CoverURL != "" {
			return
		}
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Screen Shot
	c.OnXML(`//*[@id="sscontainerppv123"]/img`, func(e *colly.XMLElement) {
		info.PreviewImages = []string{e.Request.AbsoluteURL(e.Attr("src"))}
	})

	// Preview Video
	c.OnXML(`//*[@id="player1"]/source`, func(e *colly.XMLElement) {
		info.PreviewVideoHLSURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Fields
	c.OnXML(`//*[@id="MyBody"]//div[@class="product-info-block-rev mt-20"]/div[@class="single-info"]`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span[1]`) {
		case "商品番号":
			info.Number = strings.TrimSpace(e.ChildText(`.//span[2]`))
		case "主演女優":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//span[2]`),
				(*[]string)(&info.Actors))
		case "スタジオ":
			info.Maker = strings.TrimSpace(e.ChildText(`.//span[2]`))
		case "シリーズ":
			info.Series = strings.TrimSpace(e.ChildText(`.//span[2]`))
		case "カテゴリ":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//span[2]`),
				(*[]string)(&info.Genres))
		case "発売日":
			info.ReleaseDate = parser.ParseDate(strings.Fields(e.ChildText(`.//span[2]`))[0])
		case "収録時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//span[2]`))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (ave *AVE) TidyKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

func (ave *AVE) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := ave.ClonedCollector()

	c.OnXML(`//div[@class="single-slider-product grid-view-product"]`, func(e *colly.XMLElement) {
		href := e.ChildAttr(`.//div[1]/a`, "href")
		thumb := e.ChildAttr(`.//div[1]/a/img`, "src")
		id, _ := ave.ParseIDFromURL(e.Request.AbsoluteURL(href))
		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   parserNumber(thumb),
			Title:    e.ChildText(`.//div[2]/p[@class="product-title"]/a`),
			Provider: ave.Name(),
			Homepage: e.Request.AbsoluteURL(href),
			ThumbURL: e.Request.AbsoluteURL(thumb),
			CoverURL: e.Request.AbsoluteURL(strings.ReplaceAll(thumb, "jacket_images", "bigcover")),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func parserNumber(s string) string {
	if ss := regexp.MustCompile(`(?i)/(?:dvd\d)?([a-z\d-_]+)\.jpg`).FindStringSubmatch(s); len(ss) == 2 {
		return strings.ToUpper(ss[1])
	}
	return ""
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
