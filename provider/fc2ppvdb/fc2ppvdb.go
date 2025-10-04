package fc2ppvdb

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2/fc2util"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*FC2PPVDB)(nil)

const (
	Name     = "FC2PPVDB"
	Priority = 1000 - 2
)

const (
	baseURL  = "https://fc2ppvdb.com/"
	movieURL = "https://fc2ppvdb.com/articles/%s"
)

type FC2PPVDB struct {
	*scraper.Scraper
}

func New() *FC2PPVDB {
	return &FC2PPVDB{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (fc2ppvdb *FC2PPVDB) NormalizeMovieID(id string) string {
	return fc2util.ParseNumber(id)
}

func (fc2ppvdb *FC2PPVDB) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return fc2ppvdb.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (fc2ppvdb *FC2PPVDB) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (fc2ppvdb *FC2PPVDB) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := fc2ppvdb.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("FC2-%s", id),
		Provider:      fc2ppvdb.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := fc2ppvdb.ClonedCollector()

	// Cover/Thumb Image
	c.OnXML(`//main//div[contains(@class,'container')]/div[1]/div[1]/a/img`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Title
	c.OnXML(`//main//div[contains(@class,'container')]/div[1]/div[2]/h2/a`, func(e *colly.XMLElement) {
		info.Title = e.Text
	})

	// Fields
	c.OnXML(`//main//div[contains(@class,'container')]/div[1]/div[2]/div`, func(e *colly.XMLElement) {
		if child := e.DOM.(*html.Node).FirstChild; child != nil {
			switch child.Data {
			case "ID：":
				info.ID = strings.TrimSpace(e.ChildText(`.//span`))
			case "販売者：":
				info.Maker = strings.TrimSpace(e.ChildText(`.//span`))
			case "女優：":
				parser.ParseTexts(
					htmlquery.FindOne(e.DOM.(*html.Node), `.//span`),
					(*[]string)(&info.Actors),
				)
			case "モザイク：": // mosaic
			case "販売日：":
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//span`))
			case "収録時間：":
				info.Runtime = parser.ParseRuntime(e.ChildText(`.//span`))
			case "タグ：": // tags & genres
				parser.ParseTexts(
					htmlquery.FindOne(e.DOM.(*html.Node), `.//span`),
					(*[]string)(&info.Genres),
				)
			}
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
