package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/javtube/javtube/model"
	"github.com/javtube/javtube/util"
)

var _ Provider = (*MGStage)(nil)

type MGStage struct {
	BaseURL   string
	MovieURL  string
	SearchURL string
	SampleURL string
}

func NewMGStage() Provider {
	return &MGStage{
		BaseURL:   "https://www.mgstage.com/",
		MovieURL:  "https://www.mgstage.com/product/product_detail/%s/",
		SearchURL: "https://www.mgstage.com/search/cSearch.php?search_word=%s",
		SampleURL: "https://www.mgstage.com/sampleplayer/sampleRespons.php?pid=%s",
	}
}

func (mgs *MGStage) GetMovieInfo(id string) (info *model.MovieInfo, err error) {
	info = &model.MovieInfo{
		ID:       strings.ToUpper(id),
		Homepage: fmt.Sprintf(mgs.MovieURL, strings.ToUpper(id)),
	}

	c := colly.NewCollector(extensions.RandomUserAgent)

	c.SetCookies(mgs.BaseURL, []*http.Cookie{
		{Name: "adc", Value: "1"},
	})

	// Title
	c.OnXML(`//*[@id="center_column"]/div[1]/h1`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		info.Summary = e.Attr("content")
	})

	// Thumb
	c.OnXML(`//*[@id="center_column"]/div[1]/div[1]/div/div/h2/img`, func(e *colly.XMLElement) {
		info.ThumbURL = strings.ReplaceAll(e.Attr("src"), "pf_o1", "pf_e")
	})

	// Cover
	c.OnXML(`//*[@id="EnlargeImage"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.Attr("href")
	})

	// Preview Video
	c.OnXML(`//*[@id="center_column"]/div[1]/div[1]/div/div/p[2]`, func(e *colly.XMLElement) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			data := make(map[string]string)
			if err = json.Unmarshal(r.Body, &data); err == nil {
				if url, ok := data["url"]; ok {
					re := regexp.MustCompile(`\.ism\/request?.+$`)
					info.PreviewVideoURL = re.ReplaceAllString(url, ".mp4")
				}
			}
		})
		pid := path.Base(e.ChildAttr(`.//a`, "href"))
		err = d.Visit(fmt.Sprintf(mgs.SampleURL, pid))
	})

	// Preview Images
	c.OnXML(`//*[@id="sample-photo"]/dd/ul/li`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.ChildAttr(`.//a`, "href"))
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//th`) {
		case "出演：":
			info.Actors = e.ChildTexts(`.//td/a`)
		case "メーカー：":
			info.Maker = e.ChildText(`.//td`)
		case "収録時間：":
			info.Duration = util.ParseDuration(e.ChildText(`.//td`))
		case "品番：":
			info.Number = e.ChildText(`.//td`)
		case "配信開始日：", "商品発売日：":
			if info.ReleaseDate.IsZero() {
				info.ReleaseDate = util.ParseDate(e.ChildText(`.//td`))
			}
		case "シリーズ：":
			info.Series = e.ChildText(`.//td`)
		case "レーベル：":
			info.Publisher = e.ChildText(`.//td`)
		case "ジャンル：":
			info.Tags = e.ChildTexts(`.//td/a`)
		case "評価：":
			info.Score = mgs.parseScore(e.ChildText(`.//td`))
		}
	})

	err = c.Visit(fmt.Sprintf(mgs.MovieURL, id))
	return
}

func (mgs *MGStage) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	c := colly.NewCollector(extensions.RandomUserAgent)

	c.SetCookies(mgs.BaseURL, []*http.Cookie{
		{Name: "adc", Value: "1"},
	})

	c.OnXML(`//*[@id="center_column"]/div[2]/div/ul/li`, func(e *colly.XMLElement) {
		results = append(results, &model.SearchResult{
			ID:       path.Base(e.ChildAttr(`.//h5/a`, "href")),
			Number:   path.Base(e.ChildAttr(`.//h5/a`, "href")), /* same as ID */
			Title:    strings.TrimSpace(e.ChildText(`.//a/p`)),
			ThumbURL: strings.ReplaceAll(e.ChildAttr(`.//h5/a/img`, "src"), "pf_t1", "pf_e"),
			CoverURL: strings.ReplaceAll(e.ChildAttr(`.//h5/a/img`, "src"), "pf_t1", "pb_e"),
			Score:    mgs.parseScore(e.ChildText(`.//p[@class="review"]`)),
		})
	})

	err = c.Visit(fmt.Sprintf(mgs.SearchURL, keyword))
	return
}

func (mgs *MGStage) parseScore(s string) float64 {
	s = strings.TrimSpace(strings.Fields(s)[0])
	n, _ := strconv.ParseFloat(s, 10)
	return n
}
