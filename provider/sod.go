package provider

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/util"
)

var _ Provider = (*SOD)(nil)

type SOD struct {
	BaseURL   string
	MovieURL  string
	SearchURL string
	OnTimeURL string
}

func NewSOD() Provider {
	return &SOD{
		BaseURL:   "https://ec.sod.co.jp/",
		MovieURL:  "https://ec.sod.co.jp/prime/videos/?id=%s",
		SearchURL: "https://ec.sod.co.jp/prime/videos/genre/?search_type=1&sodsearch=%s",
		OnTimeURL: "https://ec.sod.co.jp/prime/_ontime.php",
	}
}

func (sod *SOD) Name() string {
	return "SOD"
}

func (sod *SOD) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return sod.GetMovieInfoByLink(fmt.Sprintf(sod.MovieURL, url.QueryEscape(id)))
}

func (sod *SOD) GetMovieInfoByLink(link string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	info = &model.MovieInfo{
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	// ID+Number
	if ss := regexp.MustCompile(`id=(.+?)$`).FindStringSubmatch(info.Homepage); len(ss) == 2 {
		info.ID = ss[1]
		info.Number = info.ID
	}

	c := colly.NewCollector(colly.UserAgent(UA))

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", fmt.Sprintf(sod.MovieURL, url.QueryEscape(info.ID)))
	})

	// Fields
	c.OnXML(`//*[@id="v_introduction"]/tbody/tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "品番":
			info.ID = e.ChildText(`.//td[2]`)
		case "発売年月日":
			if ss := regexp.MustCompile(`([\s\d]+)年([\s\d]+)月([\s\d]+)日`).
				FindStringSubmatch(e.ChildText(`.//td[2]`)); len(ss) == 4 {
				info.ReleaseDate = util.ParseDate(fmt.Sprintf("%s-%s-%s",
					strings.TrimSpace(ss[1]), strings.TrimSpace(ss[2]), strings.TrimSpace(ss[3])))
			}
		case "シリーズ名":
			info.Series = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "出演者":
			info.Actors = e.ChildTexts(`.//td[2]/a`)
		case "再生時間":
			info.Duration = util.ParseDuration(e.ChildText(`.//td[2]`))
		case "監督":
			info.Director = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "メーカー":
			info.Maker = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "レーベル":
			info.Publisher = strings.TrimSpace(e.ChildText(`.//td[2]`))
		case "ジャンル":
			info.Tags = e.ChildTexts(`.//td[2]`)
			if tags := e.ChildTexts(`.//td[2]/a`); len(tags) > 0 {
				info.Tags = tags
			}
		}
	})

	// Title
	c.OnXML(`//p[@class="product_title"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//div[@class="videos_textli"]/article`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnXML(`//div[@class="videos_textli"]//div[@class="videos_sampb"]/a`, func(e *colly.XMLElement) {
		d := c.Clone()
		d.OnXML(`//*[@id="moviebox"]/video/source`, func(e *colly.XMLElement) {
			info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("src"))
		})
		d.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Thumb+Cover
	c.OnXML(`//*[@id="videos_toptable"]`, func(e *colly.XMLElement) {
		info.CoverURL = e.ChildAttr(`.//div[@class="videos_samimg"]/a[1]`, "href")
		info.ThumbURL = e.ChildAttr(`.//div[@class="videos_samimg"]/a[1]/img`, "src")
	})

	// Preview Images
	c.OnXML(`//*[@id="videos_samsbox"]/a`, func(e *colly.XMLElement) {
		info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(e.Attr("href")))
	})

	// Score
	c.OnXML(`//*[@id="review_body"]//div[@class="imagestar"]/i`, func(e *colly.XMLElement) {
		info.Score = util.ParseScore(e.Text)
	})

	err = c.Visit(sod.OnTimeURL)
	return
}

func (sod *SOD) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	c := colly.NewCollector(colly.UserAgent(UA))

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", fmt.Sprintf(sod.SearchURL, url.QueryEscape(keyword)))
	})

	c.OnXML(`//*[@id="videos_s_mainbox"]`, func(e *colly.XMLElement) {
		searchResult := &model.SearchResult{
			Title:    e.ChildText(`.//div[@class="videis_s_txt"]/h2/a`),
			Homepage: e.Request.AbsoluteURL(e.ChildAttr(`.//div[@class="videis_s_img"]/a`, "href")),
		}

		// ID+Number
		if ss := regexp.MustCompile(`id=(.+?)$`).FindStringSubmatch(searchResult.Homepage); len(ss) == 2 {
			searchResult.ID = ss[1]
			searchResult.Number = searchResult.ID
		}

		// Thumb+Cover
		if thumb := e.ChildAttr(`.//div[@class="videis_s_img"]/a/img`, "src"); thumb != "" {
			searchResult.ThumbURL = e.Request.AbsoluteURL(thumb)
			searchResult.CoverURL = strings.ReplaceAll(searchResult.ThumbURL, "_m.jpg", "_l.jpg")
		}

		// ReleaseDate
		if ss := regexp.MustCompile(`発売日([\s\d]+)年([\s\d]+)月([\s\d]+)日`).
			FindStringSubmatch(e.ChildText(`.//div[@class="videis_s_star"]/p`)); len(ss) == 4 {
			searchResult.ReleaseDate = util.ParseDate(
				fmt.Sprintf("%s-%s-%s",
					strings.TrimSpace(ss[1]),
					strings.TrimSpace(ss[2]),
					strings.TrimSpace(ss[3])))
		}

		results = append(results, searchResult)
	})

	err = c.Visit(sod.OnTimeURL)
	return
}
