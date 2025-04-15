package javfree

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2/fc2util"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*JAVFREE)(nil)
	_ provider.MovieSearcher = (*JAVFREE)(nil)
)

const (
	Name     = "JAVFREE"
	Priority = 1000 - 7
)

const (
	baseURL   = "https://javfree.me/"
	movieURL  = "https://javfree.me/%s/%s"
	searchURL = "https://javfree.me/?s=%s"
)

type JAVFREE struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *JAVFREE {
	return &JAVFREE{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese),
	}
}

func (javfree *JAVFREE) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	ss := strings.SplitN(id, "-", 2)
	if len(ss) != 2 {
		return nil, provider.ErrInvalidID
	}
	return javfree.GetMovieInfoByURL(fmt.Sprintf(movieURL, ss[0], "fc2-ppv-"+ss[1]))
}

func (javfree *JAVFREE) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if ss := regexp.MustCompile(`/(\d+)/fc2-ppv-(\d+)`).FindStringSubmatch(homepage.Path); len(ss) == 3 {
		return fmt.Sprintf("%s-%s", ss[1], ss[2]), nil
	}
	return "", provider.ErrInvalidURL
}

func (javfree *JAVFREE) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := javfree.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id, // Dual-ID (id+number)
		Provider:      javfree.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := javfree.ClonedCollector()

	// Title
	c.OnXML(`//header[@class="entry-header"]/h1`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(regexp.
			MustCompile(`(?i)(?:FC2(?:[-_]?PPV)?[-_]?)(\d+)`).
			ReplaceAllString(e.Text, ""))
		if num := fc2util.ParseNumber(strings.
			TrimSpace(rawURL[strings.LastIndex(rawURL, "/")+1:])); num != "" {
			info.Number = fmt.Sprintf("FC2-%s", num)
		}
	})

	// Director & Release Date
	c.OnXML(`//span[@class="post-author"]/strong`, func(e *colly.XMLElement) {
		info.Director = strings.TrimSpace(e.Text)
		if next := e.DOM.(*html.Node).NextSibling; next != nil {
			info.ReleaseDate = parser.ParseDate(next.Data)
		}
	})

	// Preview Images
	c.OnXML(`//div[@class="entry-content"]/p/img`, func(e *colly.XMLElement) {
		if href := e.Attr("src"); href != "" {
			info.PreviewImages = append(info.PreviewImages, href)
		}
	})

	// Cover (fallback)
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" && len(info.PreviewImages) > 0 {
			info.CoverURL = info.PreviewImages[0]
			info.PreviewImages = info.PreviewImages[1:]
		}
		// cover as thumb image.
		info.ThumbURL = info.CoverURL
	})

	err = c.Visit(info.Homepage)
	return
}

func (javfree *JAVFREE) NormalizeMovieKeyword(keyword string) string {
	return fc2util.ParseNumber(keyword)
}

func (javfree *JAVFREE) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := javfree.ClonedCollector()
	fc2ID := keyword[strings.LastIndex(keyword, "-")+1:]
	c.OnXML(`//article[@class="hentry clear"]`, func(e *colly.XMLElement) {
		var thumb, cover string
		thumb = e.Request.AbsoluteURL(e.ChildAttr(`.//a/div/img`, "src"))
		cover = fmt.Sprintf("https://cf.javfree.me/HLIC/%s", thumb[strings.LastIndex(thumb, "/")+1:])
		title := e.ChildText(`.//h2/a`)

		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//h2/a`, "href"))
		id, _ := javfree.ParseMovieIDFromURL(homepage)
		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   fmt.Sprintf("FC2-%s", fc2ID),
			Title:    title,
			Provider: javfree.Name(),
			Homepage: homepage,
			ThumbURL: thumb,
			CoverURL: cover,
		})
	})
	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(fc2ID)))
	return
}

func init() {
	provider.Register(Name, New)
}
