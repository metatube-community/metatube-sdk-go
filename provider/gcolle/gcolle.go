package gcolle

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*Gcolle)(nil)

const (
	Name     = "Gcolle"
	Priority = 1000
)

const (
	baseURL = "https://gcolle.net/"
	// Site revamped: old path /product_info.php/products_id/{id} 301-redirects to
	// new path /adult/contents/{id}, and the old DOM (#cart_quantity / .filesetumei / age_check) is gone.
	movieURL = "https://gcolle.net/adult/contents/%s"
)

type Gcolle struct {
	*scraper.Scraper
}

func New() *Gcolle {
	// Note: WithDetectCharset is NOT enabled — when enabled, colly mis-detects the UTF-8 response as a wrong charset,
	// causing the body to fail parsing (the page is already charset=utf-8, no detection needed).
	return &Gcolle{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
	)}
}

func (gcl *Gcolle) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:GCOLLE[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (gcl *Gcolle) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return gcl.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (gcl *Gcolle) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// Both new /adult/contents/55720 and old /product_info.php/products_id/847256 take the trailing digits
	if ss := regexp.MustCompile(`/(\d+)/?$`).FindStringSubmatch(homepage.Path); len(ss) == 2 {
		return ss[1], nil
	}
	return "", provider.ErrInvalidURL
}

// og:title carries a site suffix that must be stripped
var (
	reOGTitleSuffix  = regexp.MustCompile(`\s*[｜|]\s*アダルト動画.*$`)
	reOGTitleSuffix2 = regexp.MustCompile(`\s*[｜|]\s*Gcolle.*$`)
	// Runtime marker in the summary, e.g. "...【26:20】 980 yen"
	reRuntimeFromDesc = regexp.MustCompile(`[【\[](?:(\d+):)?(\d{1,2}):(\d{2})[】\]]`)
	// Runtime marker, e.g. "・Time:26min19sec"
	reRuntimeJa = regexp.MustCompile(`(\d+)\s*分`)
	// JSON-LD Product structure
	reLDJSON   = regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
	reCategory = regexp.MustCompile(`/adult/categories/(\d+)`)
)

func (gcl *Gcolle) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := gcl.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("GCOLLE-%s", id),
		Provider:      gcl.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := gcl.ClonedCollector()

	// Title: prefer og:title, strip site suffix
	c.OnXML(`//meta[@property="og:title"]`, func(e *colly.XMLElement) {
		title := strings.TrimSpace(e.Attr("content"))
		if title == "" {
			return
		}
		title = reOGTitleSuffix.ReplaceAllString(title, "")
		title = reOGTitleSuffix2.ReplaceAllString(title, "")
		info.Title = strings.TrimSpace(title)
	})

	// Title fallback: page h1
	c.OnXML(`//h1`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		if title := strings.TrimSpace(e.Text); title != "" {
			info.Title = title
		}
	})

	// Summary: og:description
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Attr("content"))
	})

	// Cover: og:image
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if src := strings.TrimSpace(e.Attr("content")); src != "" {
			info.CoverURL = e.Request.AbsoluteURL(src)
			info.ThumbURL = info.CoverURL
		}
	})

	// Categories: dedupe /adult/categories/{id} links
	c.OnXML(`//a[@href]`, func(e *colly.XMLElement) {
		href := e.Attr("href")
		if !reCategory.MatchString(href) {
			return
		}
		if name := strings.TrimSpace(e.Text); name != "" {
			for _, g := range info.Genres {
				if g == name {
					return
				}
			}
			info.Genres = append(info.Genres, name)
		}
	})

	// Runtime: the 【26:20】 marker at the tail of og:description
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		if info.Runtime != 0 {
			return
		}
		info.Runtime = parseRuntimeFromText(e.Attr("content"))
	})

	// JSON-LD Product: take the authoritative name / description / image
	c.OnResponse(func(r *colly.Response) {
		m := reLDJSON.FindSubmatch(r.Body)
		if len(m) != 2 {
			return
		}
		var ld struct {
			Type  string `json:"@type"`
			Name  string `json:"name"`
			Desc  string `json:"description"`
			Image any    `json:"image"`
			Brand struct {
				Name string `json:"name"`
			} `json:"brand"`
		}
		if json.Unmarshal(m[1], &ld) != nil {
			return
		}
		if info.Title == "" {
			info.Title = strings.TrimSpace(ld.Name)
		}
		if info.Summary == "" {
			info.Summary = strings.TrimSpace(ld.Desc)
		}
		if s, ok := ld.Image.(string); ok && info.CoverURL == "" {
			info.CoverURL = s
			info.ThumbURL = s
		}
		if info.Maker == "" {
			info.Maker = strings.TrimSpace(ld.Brand.Name)
		}
	})

	err = c.Visit(info.Homepage)
	if err != nil {
		return
	}

	// JSON-LD description often has details like "・Time:26min19sec"; fall back to grab the runtime
	// Note: the new Gcolle page exposes no stable release-date field, so ReleaseDate stays zero.
	if info.Runtime == 0 {
		info.Runtime = parseRuntimeFromText(info.Summary)
	}

	return
}

// parseRuntimeFromText extracts runtime (minutes) from markers like "【26:20】" / "26min19sec"
func parseRuntimeFromText(text string) int {
	// e.g. 【26:20】 or 【1:02:03】
	if m := reRuntimeFromDesc.FindStringSubmatch(text); len(m) == 4 {
		h, _ := strconv.Atoi(m[1])
		mnt, _ := strconv.Atoi(m[2])
		if h > 0 || mnt > 0 {
			return h*60 + mnt
		}
	}
	// e.g. "26min19sec"
	if m := reRuntimeJa.FindStringSubmatch(text); len(m) == 2 {
		if mnt, err := strconv.Atoi(m[1]); err == nil && mnt > 0 {
			return mnt
		}
	}
	return 0
}

func init() {
	provider.Register(Name, New)
}
