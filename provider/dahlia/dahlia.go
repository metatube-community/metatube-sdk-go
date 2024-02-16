package dahlia

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*DAHLIA)(nil)
	_ provider.MovieSearcher = (*DAHLIA)(nil)
)

const (
	Name     = "DAHLIA"
	Priority = 1000 - 1
)

const (
	baseURL   = "https://dahlia-av.jp/"
	movieURL  = "https://dahlia-av.jp/works/%s/"
	searchURL = "https://dahlia-av.jp/?s=%s"
)

type DAHLIA struct {
	*scraper.Scraper
}

func New() *DAHLIA {
	return &DAHLIA{scraper.NewDefaultScraper(Name, baseURL, Priority)}
}

func (dha *DAHLIA) NormalizeMovieID(id string) string {
	return strings.ToLower(id)
}

func (dha *DAHLIA) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return dha.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (dha *DAHLIA) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return dha.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (dha *DAHLIA) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := dha.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        id,
		Provider:      dha.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
	}

	c := dha.ClonedCollector()

	// Title
	c.OnXML(`//div[@class="bar02_works"]/h1/text()`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Cover
	c.OnXML(`//div[@class="box_works01_img"]/a/img`, func(e *colly.XMLElement) {
		info.CoverURL = strings.Split(e.Request.AbsoluteURL(e.Attr("src")), "?")[0]
	})

	// Preview Video
	c.OnXML(`//div[@class="box_works01_img"]/a`, func(e *colly.XMLElement) {
		info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Previews
	c.OnXML(`//div[@class="box_works01_ga"]//li/a`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Summary
	c.OnXML(`//div[@class="box_works01_text"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		if summary := strings.TrimSpace(e.Text); summary != "" {
			info.Summary = strings.ReplaceAll(summary, "\n", "<br />")
		}
	})

	// Fields
	c.OnXML(`//div[contains(@class, "box_works01_list")]/ul//li`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span`) {
		case "出演女優":
			info.Actors = append(info.Actors, e.ChildText(`.//p`))
		case "収録時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//p`))
		case "発売日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//p`))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (dha *DAHLIA) NormalizeMovieKeyword(keyword string) string {
	if !regexp.MustCompile(`^(?i)dldss-?\d{3}$`).MatchString(keyword) {
		return ""
	}
	return strings.ToLower(strings.ReplaceAll(keyword, "-", ""))
}

func (dha *DAHLIA) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := dha.ClonedCollector()
	c.ParseHTTPErrorResponse = true
	c.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	c.OnXML(`//div[@class="box_kanren01"]//li`, func(e *colly.XMLElement) {
		cover := e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))

		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//a`, "href"))
		id, _ := dha.ParseMovieIDFromURL(homepage)
		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      id,
			Title:       strings.SplitN(e.ChildText(`.//div[@class="text_name"]/a`), "\n", 2)[0],
			Provider:    dha.Name(),
			Homepage:    homepage,
			CoverURL:    cover,
			ReleaseDate: parser.ParseDate(strings.Fields(e.ChildText(`.//div[contains(text(), "発売開始")]`))[0]),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
