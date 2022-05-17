package fanza

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
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"golang.org/x/net/html"
)

var (
	_ provider.MovieProvider = (*FANZA)(nil)
	_ provider.MovieSearcher = (*FANZA)(nil)
)

const (
	Name     = "FANZA"
	Priority = 1000
)

const (
	baseURL                 = "https://www.dmm.co.jp/"
	baseDigitalURL          = "https://www.dmm.co.jp/digital/"
	baseMonoURL             = "https://www.dmm.co.jp/mono/"
	searchURL               = "https://www.dmm.co.jp/search/=/searchstr=%s/limit=120/sort=rankprofile/"
	movieDigitalVideoAURL   = "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=%s/"
	movieDigitalVideoCURL   = "https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=%s/"
	movieDigitalAnimeURL    = "https://www.dmm.co.jp/digital/anime/-/detail/=/cid=%s/"
	movieDigitalNikkatsuURL = "https://www.dmm.co.jp/digital/nikkatsu/-/detail/=/cid=%s/"
	movieMonoDVDURL         = "https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=%s/"
	movieMonoAnimeURL       = "https://www.dmm.co.jp/mono/anime/-/detail/=/cid=%s/"
)

type FANZA struct {
	*provider.Scraper
}

func New() *FANZA {
	c := colly.NewCollector()
	c.SetCookies(baseURL, []*http.Cookie{
		{Name: "age_check_done", Value: "1"},
	})
	return &FANZA{provider.NewScraper(Name, Priority, c)}
}

func (fz *FANZA) NormalizeID(id string) string {
	return strings.ToLower(id) // FANZA uses lowercase ID.
}

func (fz *FANZA) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	for _, homepage := range []string{
		fmt.Sprintf(movieDigitalVideoAURL, id),
		fmt.Sprintf(movieMonoDVDURL, id),
		fmt.Sprintf(movieDigitalVideoCURL, id),
		fmt.Sprintf(movieDigitalAnimeURL, id),
		fmt.Sprintf(movieMonoAnimeURL, id),
		fmt.Sprintf(movieDigitalNikkatsuURL, id),
	} {
		if info, err = fz.GetMovieInfoByURL(homepage); err == nil && info.Valid() {
			return
		}
	}
	return nil, provider.ErrNotFound
}

func (fz *FANZA) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	var id string
	if sub := regexp.MustCompile(`/cid=(.*?)/`).FindStringSubmatch(u); len(sub) == 2 {
		id = strings.ToLower(sub[1])
	} else {
		return nil, fmt.Errorf("invalid FANZA url: %s", u)
	}

	info = &model.MovieInfo{
		Provider:      fz.Name(),
		Homepage:      u,
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := fz.Collector()

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
		info.CoverURL = e.Request.AbsoluteURL(PreviewSrc(e.Attr("href")))
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "品番：":
			info.ID = e.ChildText(`.//td[2]`)
			info.Number = ParseNumber(info.ID)
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
			info.Score = fz.parseScoreFromURL(e.ChildAttr(`.//td[2]/img`, "src"))
		case "収録時間：":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td[2]`))
		case "監督：":
			info.Director = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "配信開始日：", "商品発売日：", "発売日：", "貸出開始日：":
			if time.Time(info.ReleaseDate).IsZero() {
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
			}
		}
	})

	// Actors
	c.OnXML(`//*[@id="performer"]`, func(e *colly.XMLElement) {
		parser.ParseTexts(e.DOM.(*html.Node), (*[]string)(&info.Actors))
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
			info.Number = ParseNumber(data.Sku)
			info.Title = data.Name
			info.Summary = data.Description
			info.ThumbURL = e.Request.AbsoluteURL(data.Image)
			if len(data.SubjectOf.Genre) > 0 {
				info.Tags = data.SubjectOf.Genre
			}
			if data.AggregateRating.RatingValue != "" {
				info.Score = parser.ParseScore(data.AggregateRating.RatingValue)
			}
			if data.SubjectOf.ContentUrl != "" {
				info.PreviewVideoURL = data.SubjectOf.ContentUrl
			}
		}
	})

	// Summary (fallback)
	c.OnXML(`//div[@class="mg-b20 lh4"]`, func(e *colly.XMLElement) {
		if info.Summary == "" {
			var summary string
			if summary = strings.TrimSpace(e.ChildText(`.//p[@class="mg-b20"]`)); summary != "" {
				// nop
			} else if summary = strings.TrimSpace(e.ChildText(`.//p`)); summary != "" {
				// nop
			} else if summary = strings.TrimSpace(e.Text); summary != "" {
				// nop
			}
			info.Summary = summary
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
		info.PreviewImages = append(info.PreviewImages,
			e.Request.AbsoluteURL(PreviewSrc(e.ChildAttr(`.//img`, "src"))))
	})

	// Final
	c.OnScraped(func(r *colly.Response) {
		if info.CoverURL == "" {
			// use thumb image as cover
			info.CoverURL = PreviewSrc(info.ThumbURL)
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (fz *FANZA) TidyKeyword(keyword string) string {
	if number.IsUncensored(keyword) {
		return ""
	}
	return strings.ReplaceAll(
		/* FANZA prefers lowercase */
		strings.ToLower(keyword),
		/* FANZA cannot search hyphened number */
		"-", "00")
}

func (fz *FANZA) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := fz.Collector()

	c.OnXML(`//*[@id="list"]/li`, func(e *colly.XMLElement) {
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//p[@class="tmb"]/a`, "href"))
		if !strings.HasPrefix(homepage, baseDigitalURL) && !strings.HasPrefix(homepage, baseMonoURL) {
			return // ignore other contents.
		}
		pattens := regexp.MustCompile(`/cid=(.+?)/`).FindStringSubmatch(homepage)
		if len(pattens) != 2 {
			err = errors.New("find id error")
			return
		}
		id := pattens[1]

		thumb := e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "src")
		if re := regexp.MustCompile(`(p[a-z]\.)jpg`); re.MatchString(thumb) {
			thumb = re.ReplaceAllString(thumb, "ps.jpg")
		}

		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   ParseNumber(id),
			Title:    e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "alt"),
			Provider: fz.Name(),
			Homepage: homepage,
			ThumbURL: e.Request.AbsoluteURL(thumb),
			CoverURL: e.Request.AbsoluteURL(PreviewSrc(thumb)),
			Score:    parser.ParseScore(e.ChildText(`.//p[@class="rate"]/span/span`)),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func (fz *FANZA) parseScoreFromURL(s string) float64 {
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

// ParseNumber parses FANZA-formatted id to general ID.
func ParseNumber(s string) string {
	if ss := regexp.MustCompile(`([A-Z]{2,})(\d+)`).FindStringSubmatch(strings.ToUpper(s)); len(ss) >= 3 {
		n, _ := strconv.Atoi(ss[2])
		return fmt.Sprintf("%s-%03d", ss[1], n)
	}
	return s
}

// PreviewSrc maximize the preview image.
// Ref: https://digstatic.dmm.com/js/digital/preview_jquery.js#652
// JS Code:
//// 画像パスの正規化
//function preview_src(src)
//{
//	  if (src.match(/(p[a-z]\.)jpg/)) {
//		  return src.replace(RegExp.$1, 'pl.');
//	  } else if (src.match(/consumer_game/)) {
//		  return src.replace('js-','-');
//	  } else if (src.match(/js\-([0-9]+)\.jpg$/)) {
//		  return src.replace('js-','jp-');
//	  } else if (src.match(/ts\-([0-9]+)\.jpg$/)) {
//		  return src.replace('ts-','tl-');
//	  } else if (src.match(/(\-[0-9]+\.)jpg$/)) {
//		  return src.replace(RegExp.$1, 'jp' + RegExp.$1);
//	  } else {
//		  return src.replace('-','jp-');
//	  }
//}
func PreviewSrc(s string) string {
	if re := regexp.MustCompile(`(p[a-z]\.)jpg`); re.MatchString(s) {
		return re.ReplaceAllString(s, "pl.jpg")
	} else if re = regexp.MustCompile(`consumer_game`); re.MatchString(s) {
		return strings.ReplaceAll(s, "js-", "-")
	} else if re = regexp.MustCompile(`js-(\d+)\.jpg$`); re.MatchString(s) {
		return strings.ReplaceAll(s, "js-", "jp-")
	} else if re = regexp.MustCompile(`ts-(\d+)\.jpg$`); re.MatchString(s) {
		return strings.ReplaceAll(s, "ts-", "tl-")
	} else if re = regexp.MustCompile(`(-\d+\.)jpg$`); re.MatchString(s) {
		return re.ReplaceAllString(s, "jp${1}jpg")
	} else {
		return strings.ReplaceAll(s, "-", "jp-")
	}
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
