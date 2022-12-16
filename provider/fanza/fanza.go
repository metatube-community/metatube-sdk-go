package fanza

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/imhelper"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*FANZA)(nil)
	_ provider.MovieSearcher = (*FANZA)(nil)
)

const (
	Name     = "FANZA"
	Priority = 1000 + 1
)

const (
	baseURL                 = "https://www.dmm.co.jp/"
	baseDigitalURL          = "https://www.dmm.co.jp/digital/"
	baseMonoURL             = "https://www.dmm.co.jp/mono/"
	searchURL               = "https://www.dmm.co.jp/search/=/searchstr=%s/limit=120/sort=date/"
	movieDigitalVideoAURL   = "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=%s/"
	movieDigitalVideoCURL   = "https://www.dmm.co.jp/digital/videoc/-/detail/=/cid=%s/"
	movieDigitalAnimeURL    = "https://www.dmm.co.jp/digital/anime/-/detail/=/cid=%s/"
	movieDigitalNikkatsuURL = "https://www.dmm.co.jp/digital/nikkatsu/-/detail/=/cid=%s/"
	movieMonoDVDURL         = "https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=%s/"
	movieMonoAnimeURL       = "https://www.dmm.co.jp/mono/anime/-/detail/=/cid=%s/"
)

type FANZA struct {
	*scraper.Scraper
}

func New() *FANZA {
	return &FANZA{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority,
			scraper.WithCookies(baseURL, []*http.Cookie{
				{Name: "age_check_done", Value: "1"},
			})),
	}
}

func (fz *FANZA) NormalizeID(id string) string {
	return strings.ToLower(id) /* FANZA uses lowercase ID */
}

func (fz *FANZA) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	homepages := []string{
		fmt.Sprintf(movieMonoDVDURL, id),
		fmt.Sprintf(movieDigitalVideoAURL, id),
		fmt.Sprintf(movieDigitalVideoCURL, id),
		fmt.Sprintf(movieDigitalAnimeURL, id),
		fmt.Sprintf(movieMonoAnimeURL, id),
		fmt.Sprintf(movieDigitalNikkatsuURL, id),
	}
	if regexp.MustCompile(`(?i)[a-z]+00\d{3,}`).MatchString(id) {
		// might be digital videoa url, try it first.
		homepages[0], homepages[1] = homepages[1], homepages[0]
	}
	for _, homepage := range homepages {
		if info, err = fz.GetMovieInfoByURL(homepage); err == nil && info.Valid() {
			return
		}
	}
	return nil, provider.ErrInfoNotFound
}

func (fz *FANZA) ParseIDFromURL(rawURL string) (id string, err error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if sub := regexp.MustCompile(`/cid=(.*?)/`).
		FindStringSubmatch(homepage.Path); len(sub) == 2 {
		id = fz.NormalizeID(sub[1])
	}
	return
}

func (fz *FANZA) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := fz.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		Provider:      fz.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := fz.ClonedCollector()

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
			info.Label = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "ジャンル：":
			info.Genres = e.ChildTexts(`.//td[2]/a`)
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

	// Actors (ajax)
	hasAPerformer := false
	c.OnXML(`//a[@id="a_performer"]`, func(e *colly.XMLElement) {
		hasAPerformer = true
	})

	// Actors (regular)
	c.OnXML(`//span[@id="performer"]`, func(e *colly.XMLElement) {
		parseActors(e.DOM.(*html.Node), (*[]string)(&info.Actors))
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
				info.Genres = data.SubjectOf.Genre
			}
			if data.AggregateRating.RatingValue != "" {
				info.Score = parser.ParseScore(data.AggregateRating.RatingValue)
			}
			if data.SubjectOf.ContentUrl != "" {
				info.PreviewVideoURL = data.SubjectOf.ContentUrl
			}
		}
	})

	// Title (fallback)
	c.OnXML(`//meta[@property="og:title"]`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = e.Attr("content")
	})

	// Summary (fallback)
	c.OnXML(`//div[@class="mg-b20 lh4"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		var summary string
		if summary = strings.TrimSpace(e.ChildText(`.//p[@class="mg-b20"]`)); summary != "" {
			// nop
		} else if summary = strings.TrimSpace(e.ChildText(`.//p`)); summary != "" {
			// nop
		} else {
			summary = strings.TrimSpace(e.Text)
		}
		info.Summary = summary
	})

	// Summary (incomplete fallback)
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		info.Summary = e.Attr("content")
	})

	// Thumb (fallback)
	c.OnXML(`//*[@id="sample-video"]//img`, func(e *colly.XMLElement) {
		if info.ThumbURL != "" {
			return // ignore if not empty.
		}
		if !strings.HasPrefix(e.Attr("id"), "package-src") {
			return // probably not our img.
		}
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Cover (fallback)
	c.OnXML(`//*[@id="sample-video"]//a[@name="package-image"]`, func(e *colly.XMLElement) {
		if info.CoverURL != "" {
			return // ignore if not empty.
		}
		if href := e.Attr("href"); strings.HasSuffix(href, ".jpg") {
			info.CoverURL = e.Request.AbsoluteURL(href)
		}
	})

	// Thumb (fallback again)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if info.ThumbURL != "" {
			return // ignore if not empty.
		}
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("content"))
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

	// Final (images)
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" {
			// try to convert thumb url to cover url.
			info.CoverURL = PreviewSrc(info.ThumbURL)
		}
	})

	// Final (big thumb image)
	c.OnScraped(func(_ *colly.Response) {
		if info.BigThumbURL != "" /* big thumb already exist */ ||
			info.ThumbURL == "" /* thumb url is empty */ ||
			len(info.PreviewImages) == 0 /* no preview images */ {
			return
		}

		if !strings.Contains(info.Homepage, "/digital/videoa") &&
			!strings.Contains(info.Homepage, "/mono/dvd") {
			// must be VideoA or DVD videos.
			return
		}

		if imhelper.Similar(info.ThumbURL, info.PreviewImages[0], nil) {
			// the first preview image is a big thumb image.
			info.BigThumbURL = info.PreviewImages[0]
			info.PreviewImages = info.PreviewImages[1:]
			return
		}
	})

	// Final (actors)
	c.OnScraped(func(r *colly.Response) {
		if !hasAPerformer {
			return
		}

		n, innerErr := htmlquery.Parse(bytes.NewReader(r.Body))
		if innerErr != nil {
			return
		}

		n = htmlquery.FindOne(n, `//script[contains(text(),"a#a_performer")]/text()`)
		if n == nil {
			return
		}

		if ss := regexp.MustCompile(`url:\s*'(.+?)',`).
			FindStringSubmatch(n.Data); len(ss) == 2 && strings.TrimSpace(ss[1]) != "" {
			d := c.Clone()
			d.OnXML(`/` /* root */, func(e *colly.XMLElement) {
				var actors []string
				parseActors(e.DOM.(*html.Node), &actors)
				if len(actors) > 0 {
					info.Actors = actors // replace with new actors.
				}
			})
			d.Visit(r.Request.AbsoluteURL(ss[1]))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (fz *FANZA) NormalizeKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToLower(keyword) /* FANZA prefers lowercase */
}

func (fz *FANZA) SearchMovie(keyword string) ([]*model.MovieSearchResult, error) {
	if strings.Contains(keyword, "-") {
		if results, err := fz.searchMovie(strings.Replace(keyword,
			/* FANZA cannot search hyphened number */
			"-", "00", 1) +
			/* Add a `#` sign to distinguish 001 style number */
			"#"); err == nil && len(results) > 0 {
			return results, nil
		}
	}
	// fallback to normal dvd search.
	return fz.searchMovie(strings.Replace(keyword, "-", "", 1))
}

func (fz *FANZA) searchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	defer func() {
		if err == nil && len(results) > 0 {
			r := regexp.MustCompile(`(?i)([A-Z]+)0*([1-9]*)`)
			x := r.ReplaceAllString(keyword, "${1}${2}")
			sort.SliceStable(results, func(i, j int) bool {
				a := r.ReplaceAllString(results[i].ID, "${1}${2}")
				b := r.ReplaceAllString(results[j].ID, "${1}${2}")
				return comparer.Compare(a, x) >= comparer.Compare(b, x)
			})
		}
	}()

	c := fz.ClonedCollector()

	c.OnXML(`//*[@id="list"]/li`, func(e *colly.XMLElement) {
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//p[@class="tmb"]/a`, "href"))
		if !strings.HasPrefix(homepage, baseDigitalURL) && !strings.HasPrefix(homepage, baseMonoURL) {
			return // ignore other contents.
		}
		id, _ := fz.ParseIDFromURL(homepage) // ignore error.

		thumb := e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "src")
		if re := regexp.MustCompile(`(p[a-z]\.)jpg`); re.MatchString(thumb) {
			thumb = re.ReplaceAllString(thumb, "ps.jpg")
		}

		var releaseDate string
		rate := e.ChildText(`.//p[@class="rate"]`)
		if re := regexp.MustCompile(`(配信日|発売日|貸出日)：\s*`); re.MatchString(rate) {
			releaseDate = re.ReplaceAllString(rate, "")
			rate = "" // reset rate.
		}
		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      ParseNumber(id),
			Title:       e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "alt"),
			Provider:    fz.Name(),
			Homepage:    homepage,
			ThumbURL:    e.Request.AbsoluteURL(thumb),
			CoverURL:    e.Request.AbsoluteURL(PreviewSrc(thumb)),
			Score:       parser.ParseScore(rate /* float or a dash (-) */),
			ReleaseDate: parser.ParseDate(releaseDate /* 発売日：2022/07/21 */),
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func parseActors(n *html.Node, texts *[]string) {
	if n.Type == html.TextNode {
		// custom trim function.
		if text := strings.Trim(strings.TrimSpace(n.Data), "-/"); text != "" {
			*texts = append(*texts, text)
		}
	}
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		// handle `id="a_performer"` situation.
		if n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "a_performer" {
					goto next
				}
			}
		}
		parseActors(n, texts)
	next:
		continue
	}
}

func (fz *FANZA) parseScoreFromURL(s string) float64 {
	u, err := url.Parse(s)
	if err != nil {
		return 0
	}
	gif := path.Base(u.Path)
	ext := path.Ext(gif)
	n := gif[:len(gif)-len(ext)]
	score, _ := strconv.ParseFloat(n, 64)
	if score > 5.0 {
		// Fix scores for mono/anime.
		// e.g.: https://review.dmm.com/web/images/pc/45.gif
		score = score / 10.0
	}
	return score
}

// ParseNumber parses FANZA-formatted id to general ID.
func ParseNumber(s string) string {
	if ss := regexp.MustCompile(`([A-Z]+)(\d+)`).FindStringSubmatch(strings.ToUpper(s)); len(ss) >= 3 {
		n, _ := strconv.Atoi(ss[2])
		return fmt.Sprintf("%s-%03d", ss[1], n)
	}
	return s
}

// PreviewSrc maximize the preview image.
// Ref: https://digstatic.dmm.com/js/digital/preview_jquery.js#652
// JS Code:
// // 画像パスの正規化
// function preview_src(src)
//
//	{
//		  if (src.match(/(p[a-z]\.)jpg/)) {
//			  return src.replace(RegExp.$1, 'pl.');
//		  } else if (src.match(/consumer_game/)) {
//			  return src.replace('js-','-');
//		  } else if (src.match(/js\-([0-9]+)\.jpg$/)) {
//			  return src.replace('js-','jp-');
//		  } else if (src.match(/ts\-([0-9]+)\.jpg$/)) {
//			  return src.replace('ts-','tl-');
//		  } else if (src.match(/(\-[0-9]+\.)jpg$/)) {
//			  return src.replace(RegExp.$1, 'jp' + RegExp.$1);
//		  } else {
//			  return src.replace('-','jp-');
//		  }
//	}
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
