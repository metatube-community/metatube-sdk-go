package duga

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*DUGA)(nil)
	_ provider.MovieSearcher = (*DUGA)(nil)
)

const (
	Name     = "DUGA"
	Priority = 1000 - 2
)

const (
	baseURL   = "https://duga.jp/"
	movieURL  = "https://duga.jp/ppv/%s/"
	searchURL = "https://duga.jp/search/=/q=%s/"
)

type DUGA struct {
	*scraper.Scraper
}

func New() *DUGA {
	return &DUGA{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese),
	}
}

func (duga *DUGA) NormalizeMovieID(id string) string {
	return strings.ToLower(id) // DUGA always use lowercase id.
}

func (duga *DUGA) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return duga.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (duga *DUGA) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return duga.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (duga *DUGA) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := duga.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      duga.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := duga.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="contentsname"]`, func(e *colly.XMLElement) {
		info.Title = e.Text
	})

	// Summary+JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			// Name        string `json:"name"`
			Description string `json:"description"`
		}{}
		if json.NewDecoder(strings.NewReader(e.Text)).Decode(&data) == nil {
			info.Summary = data.Description
		}
	})

	// Thumb
	c.OnXML(`//div[@class="imagebox"]//img[@id="productjpg"]`, func(e *colly.XMLElement) {
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Cover
	c.OnXML(`//div[@class="imagebox"]/a`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Fields
	c.OnXML(`//div[@class="summaryinner"]//table//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//th`) {
		case "配信開始日", "発売日":
			if time.Time(info.ReleaseDate).IsZero() {
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td`))
			}
		case "メーカー":
			info.Maker = strings.TrimSpace(e.ChildText(`.//td`))
		case "レーベル":
			info.Label = strings.TrimSpace(e.ChildText(`.//td`))
		case "作品ID":
			info.ID = strings.TrimSpace(e.ChildText(`.//td`))
		case "メーカー品番":
			info.Number = strings.TrimSpace(e.ChildText(`.//td`))
		case "シリーズ":
			info.Series = strings.TrimSpace(e.ChildText(`.//td`))
		case "出演者", "監督", "カテゴリ":
			// parse later.
		}
	})

	// Director
	c.OnXML(`//ul[@class="director"]//li//a`, func(e *colly.XMLElement) {
		if info.Director != "" {
			return // ignore others.
		}
		info.Director = e.Text
	})

	// Actors
	c.OnXML(`//ul[@class="performer"]//li//a`, func(e *colly.XMLElement) {
		info.Actors = append(info.Actors, e.Text)
	})

	// Genres
	c.OnXML(`//ul[@class="categorylist"]//li//a`, func(e *colly.XMLElement) {
		info.Genres = append(info.Genres, e.Text)
	})

	// Title (fallback)
	c.OnXML(`//meta[@property="og:title"]`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = e.Attr("content")
	})

	// Summary (fallback)
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		info.Summary = e.Attr("content")
	})

	// Thumb (fallback)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if info.ThumbURL != "" {
			return
		}
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("content"))
	})

	// Score
	c.OnXML(`//div[@class="summaryinner"]//div[@class="ratingstar-total"]`, func(e *colly.XMLElement) {
		info.Score = parser.ParseScore(e.ChildAttr(`.//img`, "alt"))
	})

	// Runtime
	c.OnXML(`//div[@class="downloadbox"]//table//tr`, func(e *colly.XMLElement) {
		if info.Runtime > 0 {
			return
		}
		if e.ChildText(`.//th`) == "再生時間" {
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td`))
		}
	})

	// Preview Video
	c.OnXML(`//video[@class="play-video"]`, func(e *colly.XMLElement) {
		info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Preview Images
	c.OnXML(`//*[@id="digestthumbbox"]/li`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages,
			e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href")))
	})

	// Multiple (fallback)
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" {
			// use thumb as cover.
			info.CoverURL = info.ThumbURL
		}
		if info.ID == "" {
			// fallback to id again.
			info.ID = id
		}
		if info.Number == "" {
			// use ID as number.
			info.Number = strings.ToUpper(info.ID)
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (duga *DUGA) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

func (duga *DUGA) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := duga.ClonedCollector()

	var ids []string
	c.OnXML(`//*[@id="searchresultarea"]//div[@class="contentslist"]`, func(e *colly.XMLElement) {
		id, _ := duga.ParseMovieIDFromURL(e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href")))
		ids = append(ids, id)
	})

	c.OnScraped(func(r *colly.Response) {
		const limit = 3
		var (
			mu sync.Mutex
			wg sync.WaitGroup
		)
		for i := 0; i < len(ids) && i < limit; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if info, _ := duga.GetMovieInfoByID(ids[i]); info != nil && info.IsValid() {
					mu.Lock()
					results = append(results, info.ToSearchResult())
					mu.Unlock()
				}
			}(i)
		}
		wg.Wait()
	})

	err = c.Visit(fmt.Sprintf(searchURL, keyword))
	return
}

func init() {
	provider.Register(Name, New)
}
