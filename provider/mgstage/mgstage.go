package mgstage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*MGS)(nil)
	_ provider.MovieSearcher = (*MGS)(nil)
	_ provider.MovieReviewer = (*MGS)(nil)
)

const (
	Name     = "MGS"
	Priority = 1000
)

const (
	baseURL   = "https://www.mgstage.com/"
	movieURL  = "https://www.mgstage.com/product/product_detail/%s/"
	searchURL = "https://www.mgstage.com/search/cSearch.php?search_word=%s"
	sampleURL = "https://www.mgstage.com/sampleplayer/sampleRespons.php?pid=%s"
)

type MGS struct {
	*scraper.Scraper
}

func New() *MGS {
	return &MGS{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		scraper.WithCookies(baseURL, []*http.Cookie{
			{Name: "adc", Value: "1"},
		}),
	)}
}

func (mgs *MGS) GetMovieReviewsByID(id string) (reviews []*model.MovieReviewDetail, err error) {
	c := mgs.ClonedCollector()

	c.OnXML(`//*[@id="user_review"]/ul/li`, func(e *colly.XMLElement) {
		name := strings.TrimSpace(regexp.MustCompile(`(さん)?(のレビュー)?`).ReplaceAllString(
			e.ChildText(`.//div[@class="user_date"]/p[@class="name"]`), ""))
		comment := strings.TrimSpace(e.ChildText(`.//p[@class="text"]`))
		if name == "" || comment == "" {
			return
		}

		score := 0.0
		stars := strings.Split(strings.TrimSpace(e.ChildAttr(`.//div[@class="user_date"]/p[@class="review"]/span`, "class")), "_")
		if len(stars) > 0 {
			score = parser.ParseScore(stars[len(stars)-1]) / 10
			if score > 5.0 {
				score = 0 // reset, must be an error
			}
		}

		reviews = append(reviews, &model.MovieReviewDetail{
			Author:  name,
			Comment: comment,
			Score:   score,
			Title:   strings.TrimSpace(e.ChildText(`.//h4`)),
			Date: parser.ParseDate(strings.ReplaceAll(
				e.ChildText(`.//p[@class="date"]`), "投稿日：", "")),
		})
	})

	err = c.Visit(fmt.Sprintf(movieURL, id))
	return
}

func (mgs *MGS) GetMovieReviewsByURL(rawURL string) (reviews []*model.MovieReviewDetail, err error) {
	id, err := mgs.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}
	return mgs.GetMovieReviewsByID(id)
}

func (mgs *MGS) NormalizeMovieID(id string) string {
	return strings.ToUpper(id)
}

func (mgs *MGS) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return mgs.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (mgs *MGS) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return mgs.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (mgs *MGS) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := mgs.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      mgs.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := mgs.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="center_column"]/div[1]/h1`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		info.Summary = e.Attr("content")
	})

	// Thumb
	c.OnXML(`//div[@class="detail_data"]/div/h2/img`, func(e *colly.XMLElement) {
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
		// Get big image from original thumb image.
		info.BigThumbURL = imageSrc(info.ThumbURL, true)
	})

	// Cover
	c.OnXML(`//*[@id="EnlargeImage"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Preview Video
	c.OnXML(`//div[@class="detail_data"]//p[@class="sample_movie_btn"]`, func(e *colly.XMLElement) {
		if pid := path.Base(e.ChildAttr(`.//a`, "href")); pid != "" {
			d := c.Clone()
			d.OnResponse(func(r *colly.Response) {
				data := make(map[string]string)
				if json.Unmarshal(r.Body, &data) == nil {
					if sample, ok := data["url"]; ok {
						info.PreviewVideoURL = regexp.MustCompile(`\.ism/request?.+$`).
							ReplaceAllString(sample, ".mp4")
					}
				}
			})
			d.Visit(fmt.Sprintf(sampleURL, pid))
		}
	})

	// Preview Images
	c.OnXML(`//*[@id="sample-photo"]/dd/ul/li`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.ChildAttr(`.//a`, "href"))
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//th`) {
		case "出演：":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `.//td`),
				(*[]string)(&info.Actors))
		case "メーカー：":
			info.Maker = e.ChildText(`.//td`)
		case "収録時間：":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td`))
		case "品番：":
			info.Number = e.ChildText(`.//td`)
		case "配信開始日：", "商品発売日：":
			if time.Time(info.ReleaseDate).IsZero() {
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td`))
			}
		case "シリーズ：":
			info.Series = e.ChildText(`.//td`)
		case "レーベル：":
			info.Label = e.ChildText(`.//td`)
		case "ジャンル：":
			info.Genres = e.ChildTexts(`.//td/a`)
		case "評価：":
			info.Score = parser.ParseScore(e.ChildText(`.//td`))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (mgs *MGS) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

func (mgs *MGS) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := mgs.ClonedCollector()

	c.OnXML(`//*[@id="center_column"]//ul[@class="product_list"]/li`, func(e *colly.XMLElement) {
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//h5/a`, "href"))
		id, _ := mgs.ParseMovieIDFromURL(homepage)
		title := strings.TrimSpace(e.ChildText(`.//a[contains(@class,'title')]`))
		if title == "" {
			title = strings.TrimSpace(e.ChildText(`.//a/p`)) // fallback
		}
		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   id, /* same as ID */
			Provider: mgs.Name(),
			Homepage: homepage,
			Title:    title,
			ThumbURL: e.Request.AbsoluteURL(imageSrc(e.ChildAttr(`.//h5/a/img`, "src"), true)),
			CoverURL: e.Request.AbsoluteURL(imageSrc(e.ChildAttr(`.//h5/a/img`, "src"), false)),
			Score:    parser.ParseScore(e.ChildText(`.//p[@class="review"]`)),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func imageSrc(s string, thumb bool) string {
	if re := regexp.MustCompile(`(?i)/p[f|b]_[a-z]\d*?_`); re.MatchString(s) {
		if thumb {
			return re.ReplaceAllString(s, "/pf_e_")
		}
		return re.ReplaceAllString(s, "/pb_e_")
	}
	return s
}

func init() {
	provider.Register(Name, New)
}
