package provider

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/util"
)

var _ Provider = (*JavBus)(nil)

type JavBus struct {
	BaseURL, MovieURL              string
	SearchURL, SearchUncensoredURL string
	ThumbURL, CoverURL             string
}

func NewJavBus() Provider {
	return &JavBus{
		BaseURL:             "https://www.javbus.com/",
		MovieURL:            "https://www.javbus.com/ja/%s",
		SearchURL:           "https://www.javbus.com/ja/search/%s",
		SearchUncensoredURL: "https://www.javbus.com/ja/uncensored/search/%s",
		ThumbURL:            "https://www.javbus.com/pics/thumb/%s%s",
		CoverURL:            "https://www.javbus.com/pics/cover/%s_b%s",
	}
}

func (bus *JavBus) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return bus.GetMovieInfoByLink(fmt.Sprintf(bus.MovieURL, strings.ToUpper(id)))
}

func (bus *JavBus) GetMovieInfoByLink(link string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		ID:            strings.ToUpper(path.Base(homepage.Path)),
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := colly.NewCollector(colly.UserAgent(UA))

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
			info.ReleaseDate = util.ParseDate(fields[len(fields)-1])
		case "収録時間:":
			fields := strings.Fields(e.Text)
			info.Duration = util.ParseDuration(fields[len(fields)-1])
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

func (bus *JavBus) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(UA),
	)

	var mu sync.Mutex
	c.OnXML(`//a[@class="movie-box"]`, func(e *colly.XMLElement) {
		mu.Lock()
		defer mu.Unlock()

		var thumb, cover string
		thumb = e.Request.AbsoluteURL(e.ChildAttr(`.//div[1]/img`, "src"))
		if re := regexp.MustCompile(`(?i)/thumbs?/([a-z\d]+)(?:_b)?\.(jpg|png)`); re.MatchString(thumb) {
			cover = re.ReplaceAllString(thumb, "/cover/${1}_b.${2}") // guess
		}

		results = append(results, &model.SearchResult{
			ID:          strings.TrimLeft(e.Attr("href"), bus.BaseURL),
			Number:      e.ChildText(`.//div[2]/span/date[1]`),
			Title:       strings.SplitN(e.ChildText(`.//div[2]/span`), "\n", 2)[0],
			Homepage:    e.Request.AbsoluteURL(e.Attr("href")),
			ThumbURL:    thumb,
			CoverURL:    cover,
			ReleaseDate: util.ParseDate(e.ChildText(`.//div[2]/span/date[2]`)),
		})
	})

	for _, u := range []string{
		fmt.Sprintf(bus.SearchURL, keyword),
		fmt.Sprintf(bus.SearchUncensoredURL, keyword)} {
		if err = c.Visit(u); err != nil {
			return nil, err
		}
	}

	c.Wait()
	return
}
