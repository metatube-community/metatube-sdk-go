package provider

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/javtube/javtube/model"
	"github.com/javtube/javtube/util"
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

func (bus *JavBus) GetMovieInfo(id string) (info *model.MovieInfo, err error) {
	info = &model.MovieInfo{
		ID:       strings.ToUpper(id),
		Homepage: fmt.Sprintf(bus.MovieURL, strings.ToUpper(id)),
	}

	c := colly.NewCollector(extensions.RandomUserAgent)

	c.OnError(func(r *colly.Response, innerErr error) {
		err = innerErr
	})

	// Image+Title
	c.OnXML(`//a[@class="bigImage"]/img`, func(e *colly.XMLElement) {
		info.Title = e.Attr("title")
		imageID, ext := bus.parseImage(e.Attr("src"))
		info.ThumbURL = e.Request.AbsoluteURL(fmt.Sprintf(bus.ThumbURL, imageID, ext))
		info.CoverURL = e.Request.AbsoluteURL(fmt.Sprintf(bus.CoverURL, imageID, ext))
	})

	// Fields
	c.OnXML(`//div[@class="col-md-3 info"]/p`, func(e *colly.XMLElement) {
		key := e.ChildText(`.//span`)

		switch key {
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
		extensions.RandomUserAgent,
	)

	c.OnXML(`//a[@class="movie-box"]`, func(e *colly.XMLElement) {
		imageID, ext := bus.parseImage(e.ChildAttr(`.//div[1]/img`, "src"))
		results = append(results, &model.SearchResult{
			ID:          strings.TrimLeft(e.Attr("href"), bus.BaseURL),
			Number:      e.ChildText(`.//div[2]/span/date[1]`),
			Title:       strings.SplitN(e.ChildText(`.//div[2]/span`), "\n", 2)[0],
			ThumbURL:    e.Request.AbsoluteURL(fmt.Sprintf(bus.ThumbURL, imageID, ext)),
			CoverURL:    e.Request.AbsoluteURL(fmt.Sprintf(bus.CoverURL, imageID, ext)),
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

func (bus *JavBus) parseImage(s string) (name, ext string) {
	if strings.Contains(s, "://") {
		u, _ := url.Parse(s)
		s = u.Path
	}
	image := path.Base(s)
	if ss := regexp.MustCompile(`^([a-zA-Z\d]+)(?:_b)?(\.\w+)$`).FindStringSubmatch(image); len(ss) == 3 {
		return ss[1], ss[2]
	}
	return "", ""
}
