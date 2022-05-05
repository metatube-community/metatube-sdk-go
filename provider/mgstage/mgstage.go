package mgstage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
)

var _ provider.Provider = (*MGStage)(nil)

const (
	baseURL   = "https://www.mgstage.com/"
	movieURL  = "https://www.mgstage.com/product/product_detail/%s/"
	searchURL = "https://www.mgstage.com/search/cSearch.php?search_word=%s"
	sampleURL = "https://www.mgstage.com/sampleplayer/sampleRespons.php?pid=%s"
)

type MGStage struct {
	c *colly.Collector
}

func NewMGStage() provider.Provider {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.IgnoreRobotsTxt(),
		colly.UserAgent(provider.UA))
	c.SetCookies(baseURL, []*http.Cookie{
		{Name: "adc", Value: "1"},
	})
	return &MGStage{c: c}
}

func (mgs *MGStage) Name() string {
	return "MGS"
}

func (mgs *MGStage) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return mgs.GetMovieInfoByLink(fmt.Sprintf(movieURL, strings.ToUpper(id)))
}

func (mgs *MGStage) GetMovieInfoByLink(link string) (info *model.MovieInfo, err error) {
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

	c := mgs.c.Clone()

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
		info.ThumbURL = e.Request.AbsoluteURL(imageSrc(e.Attr("src"), true))
	})

	// Cover
	c.OnXML(`//*[@id="EnlargeImage"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Preview Video
	c.OnXML(`//div[@class="detail_data"]//p[@class="sample_movie_btn"]`, func(e *colly.XMLElement) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			data := make(map[string]string)
			if json.Unmarshal(r.Body, &data) == nil {
				if u, ok := data["url"]; ok {
					info.PreviewVideoURL = regexp.MustCompile(`\.ism/request?.+$`).
						ReplaceAllString(u, ".mp4")
				}
			}
		})
		pid := path.Base(e.ChildAttr(`.//a`, "href"))
		d.Visit(fmt.Sprintf(sampleURL, pid))
	})

	// Preview Images
	c.OnXML(`//*[@id="sample-photo"]/dd/ul/li`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.ChildAttr(`.//a`, "href"))
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//th`) {
		case "出演：":
			if actors := e.ChildTexts(`.//td/a`); len(actors) > 0 {
				info.Actors = actors
			} else if actors = e.ChildTexts(`.//td`); len(actors) > 0 {
				for _, actor := range actors {
					info.Actors = append(info.Actors, strings.TrimSpace(actor))
				}
			}
		case "メーカー：":
			info.Maker = e.ChildText(`.//td`)
		case "収録時間：":
			info.Duration = parser.ParseDuration(e.ChildText(`.//td`))
		case "品番：":
			info.Number = e.ChildText(`.//td`)
		case "配信開始日：", "商品発売日：":
			if info.ReleaseDate.IsZero() {
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td`))
			}
		case "シリーズ：":
			info.Series = e.ChildText(`.//td`)
		case "レーベル：":
			info.Publisher = e.ChildText(`.//td`)
		case "ジャンル：":
			info.Tags = e.ChildTexts(`.//td/a`)
		case "評価：":
			info.Score = parser.ParseScore(e.ChildText(`.//td`))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (mgs *MGStage) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	c := mgs.c.Clone()

	c.OnXML(`//*[@id="center_column"]/div[2]/div/ul/li`, func(e *colly.XMLElement) {
		href := e.ChildAttr(`.//h5/a`, "href")
		results = append(results, &model.SearchResult{
			ID:       path.Base(href),
			Number:   path.Base(href), /* same as ID */
			Homepage: e.Request.AbsoluteURL(href),
			Title:    strings.TrimSpace(e.ChildText(`.//a/p`)),
			ThumbURL: e.Request.AbsoluteURL(imageSrc(e.ChildAttr(`.//h5/a/img`, "src"), true)),
			CoverURL: e.Request.AbsoluteURL(imageSrc(e.ChildAttr(`.//h5/a/img`, "src"), false)),
			Score:    parser.ParseScore(e.ChildText(`.//p[@class="review"]`)),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func imageSrc(s string, thumb bool) string {
	if re := regexp.MustCompile(`(?i)/p[f|b]_[a-z]\d+?_`); re.MatchString(s) {
		if thumb {
			return re.ReplaceAllString(s, "/pf_e_")
		}
		return re.ReplaceAllString(s, "/pb_e_")
	}
	return s
}
