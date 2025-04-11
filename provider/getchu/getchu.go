package getchu

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*Getchu)(nil)

const (
	Name     = "Getchu"
	Priority = 1000
)

const (
	baseURL  = "https://dl.getchu.com/"
	movieURL = "https://dl.getchu.com/i/item%s"
)

type Getchu struct {
	*scraper.Scraper
}

func New() *Getchu {
	return &Getchu{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (gcu *Getchu) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:GETCHU[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (gcu *Getchu) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return gcu.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (gcu *Getchu) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return strings.TrimLeft(path.Base(homepage.Path), "item"), nil
}

func (gcu *Getchu) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := gcu.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("GETCHU-%s", id),
		Provider:      gcu.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := gcu.ClonedCollector()

	// Misc
	c.OnXML(`//td`, func(e *colly.XMLElement) {
		switch {
		// Title
		case e.ChildAttr(`.//div`, "style") == "color: #333333; padding: 3px 0px 0px 5px;":
			info.Title = strings.TrimSpace(e.Text)
		// Cover
		case e.Attr("bgcolor") == "#ffffff":
			info.CoverURL = e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))
		// Preview Images
		case strings.Contains(e.Attr("style"), "background-color: #444444;"):
			info.PreviewImages = append(info.PreviewImages,
				e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href")))
		}
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "サークル":
			info.Label = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "作者":
			// info.Director = e.ChildText(`.//td[2]`)
		case "画像数&ページ数":
			// info.Runtime = parser.ParseRuntime(e.ChildText(`.//td[2]`))
		case "配信開始日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
		case "趣向":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//td[2]`),
				(*[]string)(&info.Genres))
		case "作品内容":
			info.Summary = strings.TrimSpace(e.ChildText(`.//td[2]`))
		}
	})

	// Title (fallback)
	c.OnXML(`//meta[@property="og:title"]`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = e.Attr("content")
	})

	// Cover (fallback)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if info.CoverURL != "" {
			return
		}
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("content"))
	})

	// Fallbacks
	c.OnScraped(func(_ *colly.Response) {
		if info.ThumbURL == "" {
			info.ThumbURL = info.CoverURL // same as cover.
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
