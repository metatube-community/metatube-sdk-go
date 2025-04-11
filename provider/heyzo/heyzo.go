package heyzo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/js"
	"github.com/metatube-community/metatube-sdk-go/common/m3u8"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*Heyzo)(nil)
	_ provider.MovieReviewer = (*Heyzo)(nil)
)

const (
	Name     = "HEYZO"
	Priority = 1000
)

const (
	baseURL          = "https://www.heyzo.com/"
	movieURL         = "https://www.heyzo.com/moviepages/%04s/index.html"
	sampleURL        = "https://www.heyzo.com/contents/%s/%s/%s"
	reviewPageURL    = "https://www.heyzo.com/app_v2/review_getjs/?id=%s&page=%d&r=%f&lang=%s"
	reviewShowAllURL = "https://www.heyzo.com/app_v2/review_getjs/?id=%s&showall=1&r=%f&lang=%s"
)

type Heyzo struct {
	*scraper.Scraper
}

func New() *Heyzo {
	return &Heyzo{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (hzo *Heyzo) GetMovieReviewsByID(id string) (reviews []*model.MovieReviewDetail, err error) {
	c := hzo.ClonedCollector()

	c.OnXML(`//script`, func(e *colly.XMLElement) {
		if !strings.Contains(e.Text, "reviews_get") {
			return
		}

		obj := struct {
			MovieSeq     string `json:"movie_seq"`
			Page         int    `json:"page"`
			Lang         string `json:"lang"`
			ProviderName string `json:"provider_name"`
		}{}
		if err = js.UnmarshalObject(e.Text, "object", &obj); err != nil {
			return
		}
		if obj.MovieSeq == "" {
			err = fmt.Errorf("no movie seq found on `%s`", e.Text)
			return
		}

		// Get reviews
		d := c.Clone()

		d.OnResponse(func(r *colly.Response) {
			data := struct {
				Comments []struct {
					Username string `json:"user_name"`
					Date     string `json:"date"`
					Comment  string `json:"comment"`
					Eng      string `json:"eng"`
					Score    struct {
						Overall string `json:"overall"`
					} `json:"score"`
				} `json:"comments"`
			}{}
			if err = js.UnmarshalObject(r.Body, "reviews", &data); err == nil {
				for _, row := range data.Comments {
					if row.Username == "" || row.Comment == "" {
						continue
					}
					reviews = append(reviews, &model.MovieReviewDetail{
						Author:  row.Username,
						Comment: row.Comment,
						Score:   parser.ParseScore(row.Score.Overall),
						Date:    parser.ParseDate(row.Date),
					})
				}
			}
		})

		var reviewURL string
		if obj.Page > 0 {
			reviewURL = fmt.Sprintf(reviewPageURL, obj.MovieSeq, obj.Page, rand.Float64(), obj.Lang)
		} else {
			reviewURL = fmt.Sprintf(reviewShowAllURL, obj.MovieSeq, rand.Float64(), obj.Lang)
		}
		if vErr := d.Visit(reviewURL); vErr != nil {
			err = vErr
		}
	})

	if vErr := c.Visit(fmt.Sprintf(movieURL, id)); vErr != nil {
		err = vErr
	}
	return
}

func (hzo *Heyzo) GetMovieReviewsByURL(rawURL string) (reviews []*model.MovieReviewDetail, err error) {
	id, err := hzo.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}
	return hzo.GetMovieReviewsByID(id)
}

func (hzo *Heyzo) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:heyzo[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (hzo *Heyzo) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return hzo.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (hzo *Heyzo) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(path.Dir(homepage.Path)), nil
}

func (hzo *Heyzo) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := hzo.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("HEYZO-%s", id),
		Provider:      hzo.Name(),
		Homepage:      rawURL,
		Maker:         "HEYZO",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := hzo.ClonedCollector()

	// JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Name          string `json:"name"`
			Image         string `json:"image"`
			Description   string `json:"description"`
			ReleasedEvent struct {
				StartDate string `json:"startDate"`
			} `json:"releasedEvent"`
			Video struct {
				Duration string `json:"duration"`
				Actor    string `json:"actor"`
				Provider string `json:"provider"`
			} `json:"video"`
			AggregateRating struct {
				RatingValue string `json:"ratingValue"`
			} `json:"aggregateRating"`
		}{}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			info.Title = data.Name
			info.Summary = data.Description
			info.CoverURL = e.Request.AbsoluteURL(data.Image)
			info.ThumbURL = info.CoverURL /* use cover as thumb */
			info.ReleaseDate = parser.ParseDate(data.ReleasedEvent.StartDate)
			info.Runtime = parser.ParseRuntime(data.Video.Duration)
			info.Score = parser.ParseScore(data.AggregateRating.RatingValue)
			if data.Video.Provider != "" {
				info.Maker = data.Video.Provider
			}
			if data.Video.Actor != "" {
				info.Actors = []string{data.Video.Actor}
			}
		}
	})

	// Title
	c.OnXML(`//*[@id="movie"]/h1`, func(e *colly.XMLElement) {
		if info.Title == "" {
			info.Title = strings.Fields(e.Text)[0]
		}
	})

	// Summary
	c.OnXML(`//p[@class="memo"]`, func(e *colly.XMLElement) {
		if info.Summary == "" {
			info.Summary = strings.TrimSpace(e.Text)
		}
	})

	// Thumb+Cover (fallback)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if info.CoverURL == "" {
			info.CoverURL = e.Request.AbsoluteURL(e.Attr("content"))
			info.ThumbURL = info.CoverURL
		}
	})

	// Fields
	c.OnXML(`//table[@class="movieInfo"]/tbody/tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "公開日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
		case "出演":
			info.Actors = e.ChildTexts(`.//td[2]/a/span`)
		case "シリーズ":
			info.Series = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "評価":
			info.Score = parser.ParseScore(e.ChildText(`.//span[@itemprop="ratingValue"]`))
		}
	})

	// Genres
	c.OnXML(`//ul[@class="tag-keyword-list"]`, func(e *colly.XMLElement) {
		info.Genres = e.ChildTexts(`.//li/a`)
	})

	// Video+Runtime
	c.OnXML(`//script[@type="text/javascript"]`, func(e *colly.XMLElement) {
		// Sample Video
		if strings.Contains(e.Text, "emvideo") {
			if sub := regexp.MustCompile(`emvideo = "(.+?)";`).FindStringSubmatch(e.Text); len(sub) == 2 {
				info.PreviewVideoURL = e.Request.AbsoluteURL(sub[1])
			}
		}
		// Runtime
		if strings.Contains(e.Text, "o = {") {
			if sub := regexp.MustCompile(`o = (\{.+?});`).FindStringSubmatch(e.Text); len(sub) == 2 {
				data := struct {
					Full string `json:"full"`
				}{}
				if json.Unmarshal([]byte(sub[1]), &data) == nil {
					info.Runtime = parser.ParseRuntime(data.Full)
				}
			}
		}
	})

	// Preview Video
	c.OnXML(`//*[@id="playerContainer"]/script`, func(e *colly.XMLElement) {
		if !strings.Contains(e.Text, "movieId") {
			return
		}
		var movieID, siteID string
		if sub := regexp.MustCompile(`movieId\s*=\s*'(\d+?)';`).FindStringSubmatch(e.Text); len(sub) == 2 {
			movieID = sub[1]
		}
		if sub := regexp.MustCompile(`siteID\s*=\s*'(\d+?)';`).FindStringSubmatch(e.Text); len(sub) == 2 {
			siteID = sub[1]
		}
		if movieID == "" || siteID == "" {
			return
		}
		if sub := regexp.MustCompile(`stream\s*=\s*'(.+?)'\+siteID\+'(.+?)'\+movieId\+'(.+?)';`).
			FindStringSubmatch(e.Text); len(sub) == 4 {
			d := c.Clone()
			d.OnResponse(func(r *colly.Response) {
				defer func() {
					// Sample HLS URL
					info.PreviewVideoHLSURL = r.Request.URL.String()
				}()
				if uri, _, err := m3u8.ParseBestMediaURI(bytes.NewReader(r.Body)); err == nil {
					if ss := regexp.MustCompile(`/sample/(\d+)/(\d+)/ts\.(.+?)\.m3u8`).
						FindStringSubmatch(uri); len(ss) == 4 {
						info.PreviewVideoURL = fmt.Sprintf(sampleURL, ss[1], ss[2], ss[3])
					}
				}
			})
			m3u8Link := e.Request.AbsoluteURL(fmt.Sprintf("%s%s%s%s%s", sub[1], siteID, sub[2], movieID, sub[3]))
			d.Visit(m3u8Link)
		}
	})

	// Preview Images
	c.OnXML(`//div[@class="sample-images yoxview"]/script`, func(e *colly.XMLElement) {
		for _, sub := range regexp.MustCompile(`"(/contents/.+/\d+?\.\w+?)"`).FindAllStringSubmatch(e.Text, -1) {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(sub[1]))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
