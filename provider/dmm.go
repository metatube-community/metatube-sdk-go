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
	"github.com/gocolly/colly/v2/extensions"
	"github.com/javtube/javtube/model"
	"github.com/javtube/javtube/util"
)

var _ Provider = (*DMM)(nil)

type DMM struct {
	BaseURL          string
	MovieVideoAURL   string
	MovieVideoCURL   string
	MovieAnimeURL    string
	MovieNikkatsuURL string
	SearchURL        string
}

func NewDMM() Provider {
	return &DMM{
		BaseURL:          "https://www.dmm.co.jp/",
		MovieVideoAURL:   "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=%s/",
		MovieVideoCURL:   "https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=%s/",
		MovieAnimeURL:    "https://www.dmm.co.jp/digital/anime/-/detail/=/cid=%s/",
		MovieNikkatsuURL: "https://www.dmm.co.jp/digital/nikkatsu/-/detail/=/cid=%s/",
		SearchURL:        "https://www.dmm.co.jp/digital/-/list/search/=/?searchstr=%s",
	}
}

func (dmm *DMM) GetMovieInfo(id string) (info *model.MovieInfo, err error) {
	id = strings.ToLower(id)
	info = &model.MovieInfo{
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := colly.NewCollector(extensions.RandomUserAgent)

	c.SetCookies(dmm.BaseURL, []*http.Cookie{
		{Name: "ckcy", Value: "1"},
		{Name: "age_check_done", Value: "1"},
	})

	// Homepage
	c.OnRequest(func(r *colly.Request) {
		info.Homepage = r.URL.String()
	})

	// Thumb
	c.OnXML(fmt.Sprintf(`//*[@id="package-src-%s"]`, id), func(e *colly.XMLElement) {
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Cover
	c.OnXML(fmt.Sprintf(`//*[@id="%s"]`, id), func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	// JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Name        string `json:"name"`
			Image       string `json:"image"`
			Description string `json:"description"`
			Sku         string `json:"sku"`
			SubjectOf   struct {
				ContentUrl string   `json:"contentUrl"`
				EmbedUrl   string   `json:"embedUrl"`
				Genre      []string `json:"genre"`
			} `json:"subjectOf"`
			AggregateRating struct {
				RatingValue string `json:"ratingValue"`
			} `json:"aggregateRating"`
		}{}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			info.ID = data.Sku
			info.Number = dmm.ParseNumber(data.Sku)
			info.Title = data.Name
			info.Summary = data.Description
			info.ThumbURL = data.Image
			info.Tags = data.SubjectOf.Genre
			info.Score = util.ParseScore(data.AggregateRating.RatingValue)
			if data.SubjectOf.ContentUrl == "" {
				info.PreviewVideoURL = data.SubjectOf.ContentUrl
			}
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
					if json.Unmarshal(resp[1], &data) == nil {
						sort.SliceStable(data.Bitrates, func(i, j int) bool {
							return data.Bitrates[i].Bitrate > data.Bitrates[j].Bitrate /* descending */
						})
						info.PreviewVideoURL = e.Request.AbsoluteURL(data.Bitrates[0].Src)
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

	// Actors
	c.OnXML(`//*[@id="performer"]`, func(e *colly.XMLElement) {
		info.Actors = e.ChildTexts(`.//a`)
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "品番：":
			if info.ID == "" {
				info.ID = e.ChildText(`.//td[2]`)
				info.Number = dmm.ParseNumber(info.ID)
			}
		case "シリーズ：":
			info.Series = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "メーカー：":
			info.Maker = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "レーベル：":
			info.Publisher = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "ジャンル：":
			if len(info.Tags) == 0 {
				info.Tags = e.ChildTexts(`.//td[2]/a`)
			}
		case "名前：":
			if len(info.Actors) == 0 {
				info.Actors = e.ChildTexts(`.//td[2]`)
			}
		case "平均評価：":
			if info.Score == 0 {
				info.Score = dmm.parseScoreFromURL(e.ChildAttr(`.//td[2]/img`, "src"))
			}
		case "収録時間：":
			info.Duration = util.ParseDuration(e.ChildText(`.//td[2]`))
		case "監督：":
			info.Director = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "配信開始日：", "商品発売日：", "発売日：":
			if info.ReleaseDate.IsZero() {
				info.ReleaseDate = util.ParseDate(e.ChildText(`.//td[2]`))
			}
		}
	})

	// Final
	c.OnScraped(func(r *colly.Response) {
		if info.CoverURL == "" {
			// use thumb image as cover
			info.CoverURL = info.ThumbURL
		}
	})

	for _, homePage := range []string{
		fmt.Sprintf(dmm.MovieVideoAURL, id),
		fmt.Sprintf(dmm.MovieVideoCURL, id),
		fmt.Sprintf(dmm.MovieAnimeURL, id),
		fmt.Sprintf(dmm.MovieNikkatsuURL, id),
	} {
		if err = c.Visit(homePage); err == nil && info.Valid() {
			break
		}
	}
	return
}

func (dmm *DMM) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	keyword = strings.ToLower(keyword) /* DMM prefers lowercase */
	c := colly.NewCollector(extensions.RandomUserAgent)

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
