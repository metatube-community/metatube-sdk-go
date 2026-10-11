package kin8tengoku

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*KIN8)(nil)

const (
	Name     = "KIN8"
	Priority = 1000
)

const (
	baseURL = "https://www.kin8tengoku.com/"
	// Site was rebuilt as a Next.js app; old path /moviepages/{id}/index.html 301-redirects to /movie/{id},
	// so we use the new path directly.
	movieURL = "https://www.kin8tengoku.com/movie/%s"
	// Preview images follow a fixed pattern: /{id}/pht/{n}.jpg
	// Verified on 5 movies (4057/3489/3605/4015/4052): always 1..13, 404 from the 14th, so we fix to 13 images
	previewImageURL = "https://www.kin8tengoku.com/%s/pht/%d.jpg"
	// Preview video (verified real mp4, ~8MB)
	sampleVideoURL = "https://smovie.kin8tengoku.com/%s/pht/sample.mp4"
	// Preview image count (site's fixed value)
	previewImageCount = 13
)

type KIN8 struct {
	*scraper.Scraper
}

func New() *KIN8 {
	return &KIN8{scraper.NewScraper(Name, baseURL, Priority, language.Japanese,
		scraper.WithUserAgent(browserUA),
		// Cloudflare returns 404 (masquerading as "not found") for requests missing Sec-Fetch-* headers,
		// so we must send browser navigation headers to fetch the real page.
		scraper.WithHeaders(map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"Accept-Language":           "ja,en;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
		}),
	)}
}

const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func (k8 *KIN8) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:kin8[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (k8 *KIN8) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return k8.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

var reMovieID = regexp.MustCompile(`(?i)(?:kin8[-_])?(\d+)`)

func (k8 *KIN8) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// New structure /movie/4057: last segment is the ID
	// Old structure /moviepages/4057/index.html: last segment is index.html; take the parent dir name
	seg := path.Base(homepage.Path)
	if seg == "index.html" || seg == "." || seg == "/" || seg == "" {
		seg = path.Base(path.Dir(homepage.Path))
	}
	if m := reMovieID.FindStringSubmatch(seg); len(m) == 2 {
		return m[1], nil
	}
	return "", provider.ErrInvalidURL
}

var (
	// og:title looks like "<actress> <title> | <studio>" (e.g. "Yamatonadeshiko ... | Kinkin Tengoku")
	reOGTitle = regexp.MustCompile(`^(.+?)\s*\|\s*.+$`)
	// og:description is the plot summary
	// Actor block: <div class="Movie_Detail_actor...">Model:<a href="/listpages/actor_18414_1">Name</a></div>
	reActorLink = regexp.MustCompile(`/listpages/actor_\d+`)
	// Category block: <div class="Movie_Detail_actor-type..."><a href="/listpages/538_1">Fellatio</a>...
	reGenreLink = regexp.MustCompile(`/listpages/\d+_`)
	// The embedded Next.js flight data is escaped (\") and stuffed into script tags, possibly split across multiple segments
	// self.__next_f.push([1,"..."]); matching the escaped pattern on the whole response body is safest.
	reFlightPush = regexp.MustCompile(`self\.__next_f\.push\(\[1,`)

	// Key fields of the movie object in flight data
	reFlightAct        = regexp.MustCompile(`\\"act_utf8\\":\\"(.+?)\\"`)
	reFlightName       = regexp.MustCompile(`\\"name_utf8\\":\\"(.+?)\\"`)
	reFlightMemo       = regexp.MustCompile(`\\"memo_utf8\\":\\"(.*?)\\"`)
	reFlightECPDate    = regexp.MustCompile(`\\"ecp_start_date\\":\\"\$D(\d{4}-\d{2}-\d{2})T`)
	reFlightDuration   = regexp.MustCompile(`\\"duration\\":(\d+)`)
	reFlightRating     = regexp.MustCompile(`\\"avg_rating\\":([\d.]+)`)
	reFlightCategories = regexp.MustCompile(`category_name_utf8\\":\\"(.+?)\\"(?:,[^{}]*?)?category_key_name\\":\\"(.*?)\\"`)
	// Operational tags (VIP/Sample/HD videos) are not real categories and must be dropped:
	// these items have category_key_name prefixed with op_flag_ (observed in data)
	reOpFlag = regexp.MustCompile(`^op_flag_`)
)

func (k8 *KIN8) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := k8.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}
	if id == "" || id == "." || id == "/" {
		return info, provider.ErrInvalidURL
	}

	// Normalize Homepage to the new path
	if home, e := url.Parse(fmt.Sprintf(movieURL, id)); e == nil {
		rawURL = home.String()
	}

	info = &model.MovieInfo{
		ID:              id,
		Number:          fmt.Sprintf("KIN8-%s", id),
		Provider:        k8.Name(),
		Homepage:        rawURL,
		Maker:           "金髪天國",
		Actors:          []string{},
		PreviewImages:   []string{},
		Genres:          []string{},
		CoverURL:        fmt.Sprintf(previewImageURL, id, 1),
		ThumbURL:        fmt.Sprintf(previewImageURL, id, 1),
		PreviewVideoURL: fmt.Sprintf(sampleVideoURL, id),
	}

	c := k8.ClonedCollector()

	// Title and summary: og metadata is the most stable
	c.OnXML(`//meta[@property="og:title"]`, func(e *colly.XMLElement) {
		title := strings.TrimSpace(e.Attr("content"))
		if m := reOGTitle.FindStringSubmatch(title); len(m) == 2 {
			title = strings.TrimSpace(m[1])
		}
		info.Title = title
	})

	// Summary: og:description
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Attr("content"))
	})

	// Cover: prefer og:image (same as pht/1.jpg; used here as a fallback override)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if src := strings.TrimSpace(e.Attr("content")); src != "" {
			info.CoverURL = e.Request.AbsoluteURL(src)
			info.ThumbURL = info.CoverURL
		}
	})

	// Actors: /listpages/actor_* links inside the Model block
	c.OnXML(`//div[contains(@class,"Movie_Detail_actor")]//a[@href]`, func(e *colly.XMLElement) {
		href := e.Attr("href")
		if reActorLink.MatchString(href) {
			if name := strings.TrimSpace(e.Text); name != "" {
				info.Actors = append(info.Actors, name)
			}
		}
	})

	// Categories: /listpages/{number}_ links (exclude actor_ prefix)
	c.OnXML(`//div[contains(@class,"Movie_Detail_actor-type")]//a[@href]`, func(e *colly.XMLElement) {
		href := e.Attr("href")
		if reActorLink.MatchString(href) {
			return
		}
		if reGenreLink.MatchString(href) {
			if name := strings.TrimSpace(e.Text); name != "" {
				info.Genres = append(info.Genres, name)
			}
		}
	})

	// The embedded Next.js flight data is the only source for release date / runtime / rating,
	// but it is escaped and stuffed into script tags; the Text from OnXML is already HTML-entity-decoded,
	// so we parse the raw response body directly.
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != http.StatusOK {
			return
		}
		body := string(r.Body)
		// Confirm this page contains flight data (embedded movie-detail JSON)
		if !reFlightPush.MatchString(body) {
			return
		}
		flight := body

		if info.Title == "" {
			if ss := reFlightName.FindStringSubmatch(flight); len(ss) == 2 {
				info.Title = strings.TrimSpace(ss[1])
			}
		}
		if info.Summary == "" {
			if ss := reFlightMemo.FindStringSubmatch(flight); len(ss) == 2 {
				// Restore \r\n escapes and trim surrounding whitespace
				info.Summary = strings.TrimSpace(strings.ReplaceAll(ss[1], `\r\n`, "\n"))
			}
		}
		// flight has a single act_utf8; fall back to the actor list parsed from the page when missing
		if len(info.Actors) == 0 {
			if ss := reFlightAct.FindStringSubmatch(flight); len(ss) == 2 {
				info.Actors = append(info.Actors, strings.TrimSpace(ss[1]))
			}
		}
		// Categories: category_name_utf8 must be filtered with category_key_name to drop operational tags
		if len(info.Genres) == 0 {
			for _, ss := range reFlightCategories.FindAllStringSubmatch(flight, -1) {
				if len(ss) == 3 && !reOpFlag.MatchString(ss[2]) {
					info.Genres = append(info.Genres, strings.TrimSpace(ss[1]))
				}
			}
		}
		if time.Time(info.ReleaseDate).IsZero() {
			if ss := reFlightECPDate.FindStringSubmatch(flight); len(ss) == 2 {
				info.ReleaseDate = parser.ParseDate(ss[1])
			}
		}
		if info.Runtime == 0 {
			if ss := reFlightDuration.FindStringSubmatch(flight); len(ss) == 2 {
				if sec, e := strconv.Atoi(ss[1]); e == nil && sec > 0 {
					info.Runtime = sec / 60
				}
			}
		}
		if info.Score == 0 {
			if ss := reFlightRating.FindStringSubmatch(flight); len(ss) == 2 {
				if score, e := strconv.ParseFloat(ss[1], 64); e == nil {
					info.Score = score
				}
			}
		}
	})

	err = c.Visit(info.Homepage)
	if err != nil {
		return
	}

	// Page DOM parsing and flight-data fallback may yield the same actor/category (callback order is not fixed),
	// so we dedupe in order after the visit completes.
	info.Actors = dedupStrings(info.Actors)
	info.Genres = dedupStrings(info.Genres)

	// Preview images: fill 1..13 by the site's fixed pattern
	if len(info.PreviewImages) == 0 {
		for i := 1; i <= previewImageCount; i++ {
			info.PreviewImages = append(info.PreviewImages, fmt.Sprintf(previewImageURL, id, i))
		}
	}

	return
}

// dedupStrings deduplicates in order and drops empty strings
func dedupStrings(in []string) []string {
	if len(in) < 2 {
		return in
	}
	seen := make(map[string]bool, len(in))
	uniq := in[:0]
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		uniq = append(uniq, s)
	}
	return uniq
}

func init() {
	provider.Register(Name, New)
}
