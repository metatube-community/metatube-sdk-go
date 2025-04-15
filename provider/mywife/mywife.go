package mywife

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*MyWife)(nil)

const (
	Name     = "MYWIFE"
	Priority = 1000
)

const (
	baseURL  = "https://mywife.cc/"
	movieURL = "https://mywife.cc/teigaku/model/no/%s"
)

type MyWife struct {
	*scraper.Scraper
}

func New() *MyWife {
	return &MyWife{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (mw *MyWife) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:mywife[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (mw *MyWife) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return mw.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (mw *MyWife) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (mw *MyWife) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := mw.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("MYWIFE-%s", id),
		Provider:      mw.Name(),
		Homepage:      rawURL,
		Maker:         "舞ワイフ",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := mw.ClonedCollector()

	// Title
	c.OnXML(`/html/head/title`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary (Part1)
	//c.OnXML(`//div[@class="modelsamplephototop"]/strong`, func(e *colly.XMLElement) {
	//})

	// Summary (Part2)
	c.OnXML(`//div[@class="modelsamplephototop"]/span[@class="text_overflow"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Summary (All)
	//c.OnXML(`//div[@class="modelsamplephototop"]`, func(e *colly.XMLElement) {
	//	info.Summary = strings.TrimSpace(e.Text)
	//})

	// Cover+Preview Video
	c.OnXML(`//div[@class="modelsamplephototop"]/video[@id="video"]`, func(e *colly.XMLElement) {
		if src := e.Attr("src"); src != "" {
			info.PreviewVideoURL = e.Request.AbsoluteURL(src)
		}

		if poster := e.Attr("poster"); poster != "" {
			info.CoverURL = e.Request.AbsoluteURL(poster)
		}
	})

	// Preview Images
	c.OnXML(`//div[@class="modelsamplephoto"]/div[@class="modelsample_photowaku"]`, func(e *colly.XMLElement) {
		if src := e.ChildAttr(`./img`, "src"); src != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(src))
		}
	})

	// Thumb
	c.OnScraped(func(r *colly.Response) {
		if info.CoverURL == "" {
			return
		}

		d := c.Clone()
		d.OnScraped(func(r *colly.Response) {
			info.ThumbURL = r.Request.URL.String()
			info.BigThumbURL = info.ThumbURL /* thumb is usually quality */
		})
		d.Head(strings.ReplaceAll(info.CoverURL, "topview.jpg", "thumb.jpg"))
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
