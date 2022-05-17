package javbus

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
)

var (
	_ provider.MovieProvider = (*JavBus)(nil)
	_ provider.MovieSearcher = (*JavBus)(nil)
)

const (
	Name     = "javbus"
	Priority = 1000 - 4
)

const (
	baseURL             = "https://www.javbus.com/"
	movieURL            = "https://www.javbus.com/ja/%s"
	searchURL           = "https://www.javbus.com/ja/search/%s"
	searchUncensoredURL = "https://www.javbus.com/ja/uncensored/search/%s"
)

type JavBus struct {
	*provider.Scraper
}

func New() *JavBus {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.IgnoreRobotsTxt(),
		colly.UserAgent(random.UserAgent()))
	c.SetCookies(baseURL, []*http.Cookie{
		// existmag=all
		{Name: "existmag", Value: "all"},
	})
	return &JavBus{provider.NewScraper(Name, Priority, c)}
}

func (bus *JavBus) NormalizeID(id string) string {
	return strings.ToUpper(id)
}

func (bus *JavBus) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return bus.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (bus *JavBus) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		ID:            strings.ToUpper(path.Base(homepage.Path)),
		Provider:      bus.Name(),
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := bus.Collector()

	// Image+Title
	c.OnXML(`//a[@class="bigImage"]/img`, func(e *colly.XMLElement) {
		info.Title = e.Attr("title")
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("src"))
		if re := regexp.MustCompile(`(?i)/cover/([a-z\d]+)(?:_b)?\.(jpg|png)`); re.MatchString(info.CoverURL) {
			info.ThumbURL = re.ReplaceAllString(info.CoverURL, "/thumb/${1}.${2}")
		}
	})

	// Fields
	c.OnXML(`//div[@class="col-md-3 info"]/p`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span`) {
		case "品番:":
			info.Number = e.ChildText(`.//span[2]`)
		case "発売日:":
			fields := strings.Fields(e.Text)
			info.ReleaseDate = parser.ParseDate(fields[len(fields)-1])
		case "収録時間:":
			fields := strings.Fields(e.Text)
			info.Runtime = parser.ParseRuntime(fields[len(fields)-1])
		case "監督:":
			info.Director = e.ChildText(`.//a`)
		case "メーカー:":
			info.Maker = e.ChildText(`.//a`)
		case "レーベル:":
			info.Publisher = e.ChildText(`.//a`)
		case "シリーズ:":
			info.Series = e.ChildText(`.//a`)
		}
	})

	// Tags
	c.OnXML(`//span[@class="genre"]`, func(e *colly.XMLElement) {
		if tag := strings.TrimSpace(e.ChildText(`.//label/a`)); tag != "" {
			info.Tags = append(info.Tags, tag)
		}
	})

	// Previews
	c.OnXML(`//*[@id="sample-waterfall"]/a`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Actors
	c.OnXML(`//div[@class="star-name"]`, func(e *colly.XMLElement) {
		info.Actors = append(info.Actors, e.ChildAttr(`.//a`, "title"))
	})

	err = c.Visit(info.Homepage)
	return
}

func (bus *JavBus) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	{ // pre-handle keyword
		if regexp.MustCompile(`^(?i)FC2-`).MatchString(keyword) {
			return nil, provider.ErrInvalidKeyword
		}
		keyword = strings.ToUpper(keyword)
	}

	c := bus.Collector()
	c.Async = true /* ASYNC */

	var mu sync.Mutex
	c.OnXML(`//a[@class="movie-box"]`, func(e *colly.XMLElement) {
		mu.Lock()
		defer mu.Unlock()

		var thumb, cover string
		thumb = e.Request.AbsoluteURL(e.ChildAttr(`.//div[1]/img`, "src"))
		if re := regexp.MustCompile(`(?i)/thumbs?/([a-z\d]+)(?:_b)?\.(jpg|png)`); re.MatchString(thumb) {
			cover = re.ReplaceAllString(thumb, "/cover/${1}_b.${2}") // guess
		}

		results = append(results, &model.MovieSearchResult{
			ID:          strings.TrimLeft(e.Attr("href"), baseURL),
			Number:      e.ChildText(`.//div[2]/span/date[1]`),
			Title:       strings.SplitN(e.ChildText(`.//div[2]/span`), "\n", 2)[0],
			Provider:    bus.Name(),
			Homepage:    e.Request.AbsoluteURL(e.Attr("href")),
			ThumbURL:    thumb,
			CoverURL:    cover,
			ReleaseDate: parser.ParseDate(e.ChildText(`.//div[2]/span/date[2]`)),
		})
	})

	for _, u := range []string{
		fmt.Sprintf(searchURL, keyword),
		fmt.Sprintf(searchUncensoredURL, keyword)} {
		if err = c.Visit(u); err != nil {
			return nil, err
		}
	}
	c.Wait()
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
