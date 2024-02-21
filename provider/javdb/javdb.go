package javdb

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*JavDB)(nil)
	_ provider.MovieSearcher = (*JavDB)(nil)
)

const (
	Name     = "JavDB"
	Priority = 1000 - 6
)

const (
	baseURL   = "https://javdb.com/"
	movieURL  = "https://javdb.com/v/%s"
	searchURL = "https://javdb.com/search?q=%s&f=all"
)

type JavDB struct {
	*scraper.Scraper
}

func (db *JavDB) NormalizeMovieKeyword(Keyword string) string {
	return Keyword
}

func New() *JavDB {
	return &JavDB{scraper.NewDefaultScraper(Name, baseURL, Priority)}
}

func (db *JavDB) NormalizeID(id string) string {
	return id
}

func (db *JavDB) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return db.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (db *JavDB) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return db.NormalizeID(path.Base(homepage.Path)), nil
}

func (db *JavDB) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := db.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      db.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := db.ClonedCollector()
	// Title
	c.OnXML(`//strong[@class="current-title"]`, func(e *colly.XMLElement) {
		info.Title = e.Text
	})

	// Image
	c.OnXML(`//div[@class="column column-video-cover"]/a/img`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("src"))
		info.ThumbURL = info.CoverURL
	})

	// Fields
	c.OnXML(`//nav[@class="panel movie-panel-info"]/div`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//strong`) {
		case "番號:":
			//fmt.Printf("番號: %s\n", e.ChildText(`.//span[1]`))
			info.Number = e.ChildText(`.//span[1]`)

		case "日期:":
			//fmt.Printf("日期: %s\n", e.ChildText(`.//span[1]`))
			fields := strings.Fields(e.ChildText(`.//span[1]`))
			info.ReleaseDate = parser.ParseDate(fields[len(fields)-1])

		case "片商:":
			//fmt.Printf("片商-Mark: %s\n", e.ChildText(`.//span[1]/a`))
			info.Maker = e.ChildText(`.//span[1]/a`)

		case "系列:":
			//fmt.Printf("系列:Series: %s\n", e.ChildText(`.//span[1]/a`))
			info.Series = e.ChildText(`.//span[1]/a`)
		// Genres
		case "類別:":
			//fmt.Printf("Genres: %s\n", genres)
			var genres = e.ChildTexts(`.//span[@class="value"]/a`)
			info.Genres = append(info.Genres, genres...)

		// Actors
		case "演員:":
			//fmt.Printf("演員: %s\n", actors)
			var actors = e.ChildTexts(`.//span[@class="value"]/a`)
			info.Actors = append(info.Actors, actors...)

		}
	})

	// Previews
	c.OnXML(`//div[@class="tile-images preview-images"]/a`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("href")))
		//fmt.Printf("PreviewImages: %s\n", e.Request.AbsoluteURL(e.Attr("href")))
	})

	err = c.Visit(info.Homepage)
	return
}

func (db *JavDB) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := db.ClonedCollector()
	c.Async = true /* ASYNC */

	var mu sync.Mutex
	c.OnXML(`//div[@class="movie-list h cols-4 vcols-8"]/div[position() < 4][@class="item"]/a`, func(e *colly.XMLElement) {
		mu.Lock()
		defer mu.Unlock()
		homepage := e.Request.AbsoluteURL(e.Attr("href"))

		//TODO 直接爬取简略信息， 不用去详细列表中获取, 暂不优化
		var info *model.MovieInfo
		if info, err = db.GetMovieInfoByURL(homepage); err != nil {
			return
		}
		results = append(results, info.ToSearchResult())
	})

	keyword = regexp.MustCompile(`\.\d{2}\.\d{2}\.\d{2}`).ReplaceAllString(keyword, "")

	for _, u := range []string{
		fmt.Sprintf(searchURL, keyword),
	} {
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
