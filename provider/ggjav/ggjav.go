package ggjav

import (
	"fmt"
	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*GGJAV)(nil)
	_ provider.MovieSearcher = (*GGJAV)(nil)
)

const (
	Name     = "GGJAV"
	Priority = 1000 - 1
)

const (
	baseURL   = "https://ggjav.com/"
	movieURL  = "https://ggjav.com/main/video?id=%s"
	searchURL = "https://ggjav.com/main/search?string=%s"
)

type GGJAV struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *GGJAV {
	return &GGJAV{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL, SkipVerify: true}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority),
	}
}

func (ggjav *GGJAV) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	ss := strings.SplitN(id, "-", 2)
	if len(ss) != 2 {
		return nil, provider.ErrInvalidID
	}
	return ggjav.GetMovieInfoByURL(fmt.Sprintf(movieURL, ss[0]))
}

func (ggjav *GGJAV) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// 提取查询参数
	queryParams := homepage.Query()
	id := queryParams.Get("id")

	// 输出结果
	if id != "" {
		return id, nil
	}
	return "", provider.ErrInvalidURL
}

func (ggjav *GGJAV) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := ggjav.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id, // Dual-ID (id+number)
		Provider:      ggjav.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := ggjav.ClonedCollector()

	// Number
	c.OnXML(`//div[@class="columns large-6 medium-4"]/div[1]`, func(e *colly.XMLElement) {
		runes := []rune(strings.TrimSpace(e.Text)) // 将字符串转换为rune切片（按字符处理）

		if num := fc2.ParseNumber(strings.TrimSpace(string(runes[3:]))); num != "" {
			info.ID = fmt.Sprintf("%s-%s", id, num)
			info.Number = fmt.Sprintf("FC2-%s", num)
		}
	})

	// Genres
	c.OnXML(`//div[@class="blue_button button ctg_button"]`, func(e *colly.XMLElement) {
		if genre := strings.TrimSpace(e.Text); genre != "" {
			info.Genres = append(info.Genres, genre)
		}
	})

	// Cover
	c.OnXML(`//div[@class="columns large-6 medium-8"]/img`, func(e *colly.XMLElement) {
		regex := `(?i)(?:FC2(?:[-_]?PPV)?[-_]?)(\d+)`
		re, _ := regexp.Compile(regex)
		replacedStr := re.ReplaceAllString(e.Attr("alt"), "")
		info.Title = strings.TrimSpace(replacedStr)
		info.CoverURL = e.Attr("src")
		// cover as thumb image.
		info.ThumbURL = info.CoverURL
	})

	err = c.Visit(info.Homepage)
	return
}

func (ggjav *GGJAV) NormalizeMovieKeyword(keyword string) string {
	return fc2.ParseNumber(keyword)
}

func (ggjav *GGJAV) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := ggjav.ClonedCollector()
	fc2ID := keyword[strings.LastIndex(keyword, "-")+1:]
	c.OnXML(`//div[@class="columns large-3 medium-6 small-12 item float-left;"]`, func(e *colly.XMLElement) {

		var thumb, cover string
		// 提取图片地址
		thumb = e.ChildAttr(`.//img[@class="item_image"]`, "src")
		cover = strings.ReplaceAll(thumb, "small", "large")
		// 提取标题
		title := e.ChildText(`.//div[@class="item_title"]/a`)

		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//div[@class="item_title"]/a`, "href"))
		id, _ := ggjav.ParseMovieIDFromURL(homepage)
		results = append(results, &model.MovieSearchResult{
			ID:       fmt.Sprintf("%s-%s", id, fc2ID),
			Number:   fmt.Sprintf("FC2-PPV-%s", fc2ID),
			Title:    title,
			Provider: ggjav.Name(),
			Homepage: homepage,
			ThumbURL: thumb,
			CoverURL: cover,
		})
	})
	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(fc2ID)))
	return
}

func init() {
	provider.Register(Name, New)
}
