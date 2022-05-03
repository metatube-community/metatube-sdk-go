package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/util"
)

var _ Provider = (*DMM)(nil)

type DMM struct {
	BaseURL                 string
	SearchURL               string
	MovieDigitalVideoAURL   string
	MovieDigitalVideoCURL   string
	MovieDigitalAnimeURL    string
	MovieDigitalNikkatsuURL string
	MovieMonoDVDURL         string
	MovieMonoAnimeURL       string
}

func NewDMM() Provider {
	return &DMM{
		BaseURL:                 "https://www.dmm.co.jp/",
		SearchURL:               "https://www.dmm.co.jp/digital/-/list/search/=/?searchstr=%s",
		MovieDigitalVideoAURL:   "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=%s/",
		MovieDigitalVideoCURL:   "https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=%s/",
		MovieDigitalAnimeURL:    "https://www.dmm.co.jp/digital/anime/-/detail/=/cid=%s/",
		MovieDigitalNikkatsuURL: "https://www.dmm.co.jp/digital/nikkatsu/-/detail/=/cid=%s/",
		MovieMonoDVDURL:         "https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=%s/",
		MovieMonoAnimeURL:       "https://www.dmm.co.jp/mono/anime/-/detail/=/cid=%s/",
	}
}

func (dmm *DMM) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	for _, homepage := range []string{
		fmt.Sprintf(dmm.MovieDigitalVideoAURL, id),
		fmt.Sprintf(dmm.MovieMonoDVDURL, id),
		fmt.Sprintf(dmm.MovieDigitalVideoCURL, id),
		fmt.Sprintf(dmm.MovieDigitalAnimeURL, id),
		fmt.Sprintf(dmm.MovieMonoAnimeURL, id),
		fmt.Sprintf(dmm.MovieDigitalNikkatsuURL, id),
	} {
		if info, err = dmm.GetMovieInfoByLink(homepage); err == nil && info.Valid() {
			return
		}
	}
	return nil, errors.New(http.StatusText(http.StatusNotFound))
}

func (dmm *DMM) GetMovieInfoByLink(link string) (info *model.MovieInfo, err error) {
	var id string
	if sub := regexp.MustCompile(`/cid=(.*?)/`).FindStringSubmatch(link); len(sub) == 2 {
		id = strings.ToLower(sub[1])
	} else {
		return nil, fmt.Errorf("invalid DMM link: %s", link)
	}

	info = &model.MovieInfo{
		Homepage:      link,
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := colly.NewCollector(colly.UserAgent(UA))

	c.SetCookies(dmm.BaseURL, []*http.Cookie{
		{Name: "ckcy", Value: "1"},
		{Name: "age_check_done", Value: "1"},
	})

	// Homepage
	c.OnRequest(func(r *colly.Request) {
		info.Homepage = r.URL.String()
	})

	// Title
	c.OnXML(`//*[@id="title"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Thumb
	c.OnXML(fmt.Sprintf(`//*[@id="package-src-%s"]`, id), func(e *colly.XMLElement) {
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Cover
	c.OnXML(fmt.Sprintf(`//*[@id="%s"]`, id), func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "品番：":
			info.ID = e.ChildText(`.//td[2]`)
			info.Number = dmm.ParseNumber(info.ID)
		case "シリーズ：":
			info.Series = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "メーカー：":
			info.Maker = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "レーベル：":
			info.Publisher = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "ジャンル：":
			info.Tags = e.ChildTexts(`.//td[2]/a`)
		case "名前：":
			info.Actors = e.ChildTexts(`.//td[2]`)
		case "平均評価：":
			info.Score = dmm.parseScoreFromURL(e.ChildAttr(`.//td[2]/img`, "src"))
		case "収録時間：":
			info.Duration = util.ParseDuration(e.ChildText(`.//td[2]`))
		case "監督：":
			info.Director = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "配信開始日：", "商品発売日：", "発売日：":
			info.ReleaseDate = util.ParseDate(e.ChildText(`.//td[2]`))
		}
	})

	// Actors
	c.OnXML(`//*[@id="performer"]`, func(e *colly.XMLElement) {
		info.Actors = e.ChildTexts(`.//a`)
	})

	// JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Name        string `json:"name"`
			Image       string `json:"image"`
			Description string `json:"description"`
			Sku         string `json:"sku"`
			SubjectOf   struct {
				ContentUrl string `json:"contentUrl"`
				// EmbedUrl   string   `json:"embedUrl"`
				Genre []string `json:"genre"`
			} `json:"subjectOf"`
			AggregateRating struct {
				RatingValue string `json:"ratingValue"`
			} `json:"aggregateRating"`
		}{ /* assign default values */
			Name:        info.Title,
			Image:       info.ThumbURL,
			Description: info.Summary,
			Sku:         info.ID,
		}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			info.ID = data.Sku
			info.Number = dmm.ParseNumber(data.Sku)
			info.Title = data.Name
			info.Summary = data.Description
			info.ThumbURL = e.Request.AbsoluteURL(data.Image)
			if len(data.SubjectOf.Genre) > 0 {
				info.Tags = data.SubjectOf.Genre
			}
			if data.AggregateRating.RatingValue != "" {
				info.Score = util.ParseScore(data.AggregateRating.RatingValue)
			}
			if data.SubjectOf.ContentUrl != "" {
				info.PreviewVideoURL = data.SubjectOf.ContentUrl
			}
		}
	})

	// Summary (fallback)
	c.OnXML(`//div[@class="mg-b20 lh4"]`, func(e *colly.XMLElement) {
		if info.Summary == "" {
			if summary := e.ChildText(`.//p`); summary != "" {
				info.Summary = strings.TrimSpace(summary)
				return
			}
			info.Summary = strings.TrimSpace(e.Text)
		}
	})

	// Summary (incomplete fallback)
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		if info.Summary == "" {
			info.Summary = e.Attr("content")
		}
	})

	// Preview Video
	c.OnXML(`//*[@id="detail-sample-movie"]/div/a`, func(e *colly.XMLElement) {
		d := c.Clone()
		d.OnXML(`//iframe`, func(e *colly.XMLElement) {
			d.OnResponse(func(r *colly.Response) {
				if resp := regexp.MustCompile(`const args = (\{.+});`).FindSubmatch(r.Body); len(resp) == 2 {
					data := struct {
						Bitrates []struct {
							Bitrate int    `json:"bitrate"`
							Src     string `json:"src"`
						} `json:"bitrates"`
					}{}
					if json.Unmarshal(resp[1], &data) == nil && len(data.Bitrates) > 0 {
						sort.SliceStable(data.Bitrates, func(i, j int) bool {
							return data.Bitrates[i].Bitrate < data.Bitrates[j].Bitrate
						})
						info.PreviewVideoURL = e.Request.AbsoluteURL(data.Bitrates[len(data.Bitrates)-1].Src)
					}
				}
			})
			d.Visit(e.Request.AbsoluteURL(e.Attr("src")))
		})
		d.Visit(e.Request.AbsoluteURL(regexp.MustCompile(`/(.+)/`).
			FindString(e.Attr("onclick"))))
	})

	// Preview Video (VR)
	c.OnXML(`//*[@id="detail-sample-vr-movie"]/div/a`, func(e *colly.XMLElement) {
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			sub := regexp.MustCompile(`var sampleUrl = "(.+?)";`).FindSubmatch(r.Body)
			if len(sub) == 2 {
				info.PreviewVideoURL = e.Request.AbsoluteURL(string(sub[1]))
			}
		})
		d.Visit(e.Request.AbsoluteURL(regexp.MustCompile(`/(.+)/`).
			FindString(e.Attr("onclick"))))
	})

	// Preview Images
	c.OnXML(`//*[@id="sample-image-block"]/a`, func(e *colly.XMLElement) {
		if image := regexp.
			MustCompile(fmt.Sprintf(`^(.+/%s).*?(-\d+\.\w+)$`, id)).
			ReplaceAllString(e.ChildAttr(`.//img`, "src"), "${1}jp${2}"); image != "" {
			info.PreviewImages = append(info.PreviewImages, image)
		} else /* fallback */ {
			info.PreviewImages = append(info.PreviewImages,
				regexp.
					MustCompile(`^(.+?)(?:js)?(-\d+\.\w+)$`).
					ReplaceAllString(e.ChildAttr(`.//img`, "src"), "${1}jp${2}"))
		}
	})

	// Final
	c.OnScraped(func(r *colly.Response) {
		if info.CoverURL == "" {
			// use thumb image as cover
			info.CoverURL = info.ThumbURL
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (dmm *DMM) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	keyword = strings.ToLower(keyword) /* DMM prefers lowercase */
	c := colly.NewCollector(colly.UserAgent(UA))

	c.SetCookies(dmm.BaseURL, []*http.Cookie{
		{Name: "ckcy", Value: "1"},
		{Name: "age_check_done", Value: "1"},
	})

	c.OnXML(`//*[@id="list"]/li`, func(e *colly.XMLElement) {
		pattens := regexp.
			MustCompile(`/cid=(.+?)/`).
			FindStringSubmatch(e.ChildAttr(`.//p[@class="tmb"]/a`, "href"))
		if len(pattens) != 2 {
			err = errors.New("find id error")
			return
		}
		id := pattens[1]
		thumb := e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "src")

		results = append(results, &model.SearchResult{
			ID:       id,
			Number:   dmm.ParseNumber(id),
			Title:    e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "alt"),
			Homepage: e.Request.AbsoluteURL(e.ChildAttr(`.//p[@class="tmb"]/a`, "href")),
			ThumbURL: e.Request.AbsoluteURL(strings.ReplaceAll(thumb, "pt.", "ps.")),
			CoverURL: e.Request.AbsoluteURL(strings.ReplaceAll(thumb, "pt.", "pl.")),
			Score:    util.ParseScore(e.ChildText(`.//p[@class="rate"]/span/span`)),
		})
	})

	err = c.Visit(fmt.Sprintf(dmm.SearchURL, keyword))
	return
}

func (dmm *DMM) ParseNumber(s string) string {
	s = strings.ToUpper(s)
	if ss := regexp.MustCompile(`([A-Z]{2,})(\d+)`).FindStringSubmatch(s); len(ss) >= 3 {
		n, _ := strconv.Atoi(ss[2])
		return fmt.Sprintf("%s-%03d", ss[1], n)
	}
	return ""
}

func (dmm *DMM) parseScoreFromURL(s string) float64 {
	u, err := url.Parse(s)
	if err != nil {
		return 0
	}
	gif := path.Base(u.Path)
	ext := path.Ext(gif)
	n := gif[:len(gif)-len(ext)]
	score, _ := strconv.ParseFloat(n, 10)
	return score
}
