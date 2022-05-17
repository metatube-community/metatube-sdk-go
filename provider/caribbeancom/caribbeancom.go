package caribbeancom

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
	"golang.org/x/net/html"
)

var _ provider.MovieProvider = (*Caribbeancom)(nil)

const (
	Name     = "CARIBBEANCOM"
	Priority = 1000
)

const (
	baseURL  = "https://www.caribbeancom.com/"
	movieURL = "https://www.caribbeancom.com/moviepages/%s/index.html"
)

type Caribbeancom struct {
	*scraper.Scraper
	DefaultMaker string
}

func New() *Caribbeancom {
	return &Caribbeancom{
		Scraper:      scraper.NewDefaultScraper(Name, Priority, scraper.WithDetectCharset()),
		DefaultMaker: "カリビアンコム",
	}
}

func (carib *Caribbeancom) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}-\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func (carib *Caribbeancom) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return carib.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (carib *Caribbeancom) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	id := path.Base(path.Dir(homepage.Path))

	info = &model.MovieInfo{
		ID:            id,
		Number:        id,
		Provider:      carib.Name(),
		Homepage:      homepage.String(),
		Maker:         carib.DefaultMaker,
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := carib.ClonedCollector()

	// Title
	c.OnXML(`//h1[@itemprop="name"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//p[@itemprop="description"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Title+Summary (Fallback)
	c.OnXML(`//div[@id="moviepages"]`, func(e *colly.XMLElement) {
		if info.Title == "" {
			info.Title = strings.TrimSpace(e.ChildText(`.//h1[1]`))
		}
		if info.Summary == "" {
			info.Summary = strings.TrimSpace(e.ChildText(`.//p[1]`))
		}
	})

	// Fields
	c.OnXML(`//*[@id="moviepages"]//li`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//span[1]`) {
		case "出演":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//span[2]`),
				(*[]string)(&info.Actors))
		case "配信日", "販売日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//span[2]`))
		case "再生時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//span[2]`))
		case "シリーズ":
			info.Series = e.ChildText(`.//span[2]/a[1]`)
		case "スタジオ":
			info.Maker /* studio */ = e.ChildText(`.//span[2]/a[1]`)
		case "タグ":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//span[2]`),
				(*[]string)(&info.Tags))
		case "ユーザー評価":
			info.Score = float64(utf8.RuneCountInString(
				strings.TrimSpace(e.ChildText(`.//span[2]`))))
		}
	})

	// Thumb+Cover+Video
	c.OnXML(`//script`, func(e *colly.XMLElement) {
		if re := regexp.MustCompile(`emimg\s*=\s*'(.+?)';`); re.MatchString(e.Text) {
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 2 {
				info.ThumbURL = e.Request.AbsoluteURL(ss[1])
				info.CoverURL = info.ThumbURL /* use thumb as cover */
			}
		} else if re = regexp.MustCompile(`posterImage\s*=\s*'(.+?)'\+movie_id\+'(.+?)';`); re.MatchString(e.Text) {
			// var posterImage = '/moviepages/'+movie_id+'/images/main_b.jpg';
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 3 {
				info.ThumbURL = e.Request.AbsoluteURL(ss[1] + id + ss[2])
				info.CoverURL = info.ThumbURL /* use thumb as cover */
			}
		} else if re = regexp.MustCompile(`Movie\s*=\s*(\{.+?});`); re.MatchString(e.Text) {
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 2 {
				data := struct {
					SampleFlashURL  string `json:"sample_flash_url"`
					SampleMFlashURL string `json:"sample_m_flash_url"`
				}{}
				if json.Unmarshal([]byte(ss[1]), &data) == nil {
					for _, sample := range []string{
						data.SampleFlashURL, data.SampleMFlashURL,
					} {
						if sample != "" {
							info.PreviewVideoURL = e.Request.AbsoluteURL(sample)
							break
						}
					}
				}
			}
		}
	})

	// Preview Images
	c.OnXML(`//div[@class="gallery-ratio"]/a`, func(e *colly.XMLElement) {
		if href := e.Attr("href"); !strings.Contains(href, "member") {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(href))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
