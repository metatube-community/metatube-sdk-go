package fantia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*Fantia)(nil)

const (
	Name     = "Fantia"
	Priority = 1000
)

const (
	baseURL        = "https://fantia.jp/"
	productBaseURL = baseURL + "products/"
	movieURL       = productBaseURL + "%s"
)

type Fantia struct {
	*scraper.Scraper
	sessionID string
}

func New() *Fantia {
	return &Fantia{Scraper: scraper.NewDefaultScraper(
		Name, productBaseURL, Priority,
		language.Japanese,
		scraper.WithCookies(baseURL, []*http.Cookie{
			{Name: "age_check", Value: "1"},
		}),
	), sessionID: strings.TrimSpace(os.Getenv("FANTIA_SESSION_ID"))}
}

func (ft *Fantia) SetConfig(c provider.Config) error {
	if !c.Has("session_id") {
		return nil
	}
	v, err := c.GetString("session_id")
	if err != nil {
		return err
	}
	ft.sessionID = strings.TrimSpace(v)
	return nil
}

func (ft *Fantia) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:FANTIA[-_])?(\d{1,})$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (ft *Fantia) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	id = ft.NormalizeMovieID(id)
	if id == "" {
		return nil, fmt.Errorf("invalid Fantia product ID")
	}
	return ft.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (ft *Fantia) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if homepage.Scheme != "https" || !strings.EqualFold(homepage.Host, "fantia.jp") {
		return "", fmt.Errorf("invalid Fantia product URL: %q", rawURL)
	}
	parts := strings.Split(strings.Trim(homepage.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "products" {
		return "", fmt.Errorf("invalid Fantia product URL: %q", rawURL)
	}
	id := ft.NormalizeMovieID(parts[1])
	if id == "" {
		return "", fmt.Errorf("invalid Fantia product URL: %q", rawURL)
	}
	return id, nil
}

func (ft *Fantia) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := ft.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}
	homepage := fmt.Sprintf(movieURL, id)

	info = &model.MovieInfo{
		ID:       id,
		Number:   fmt.Sprintf("FANTIA-%s", id),
		Provider: ft.Name(),
		Homepage: homepage,
	}

	c := ft.ClonedCollector()

	if ft.sessionID != "" {
		if err := c.SetCookies(info.Homepage, []*http.Cookie{
			{Name: "_session_id", Value: ft.sessionID, Secure: true, HttpOnly: true, Path: "/"},
		}); err != nil {
			return nil, fmt.Errorf("set Fantia session cookie: %w", err)
		}
	}

	var gtmTags []string

	// Title
	c.OnHTML(`h1.product-title`, func(e *colly.HTMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Maker
	c.OnHTML(`h1.fanclub-name a`, func(e *colly.HTMLElement) {
		info.Maker = strings.TrimSpace(e.Text)
	})

	// Description
	c.OnHTML(`div.product-description`, func(e *colly.HTMLElement) {
		info.Summary = longerSummary(info.Summary, productDescription(e.DOM))
	})
	c.OnHTML(`meta[property="og:description"]`, func(e *colly.HTMLElement) {
		info.Summary = longerSummary(info.Summary, e.Attr("content"))
	})

	// Cover
	c.OnHTML(`meta[property="og:image"]`, func(e *colly.HTMLElement) {
		info.CoverURL = e.Attr("content")
		info.ThumbURL = info.CoverURL
	})

	// Preview Images
	c.OnHTML(`div.product-gallery img`, func(e *colly.HTMLElement) {
		if src := e.Attr("src"); src != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(src))
		}
	})

	// JSON-LD
	c.OnHTML(`script[type="application/ld+json"]`, func(e *colly.HTMLElement) {
		for _, item := range decodeSchemaMovies(e.Text) {
			if item.Type != "Product" && item.Type != "VideoObject" {
				continue
			}
			if info.Title == "" {
				info.Title = strings.TrimSpace(item.Name)
			}
			info.Summary = longerSummary(info.Summary, item.Description)
			if time.Time(info.ReleaseDate).IsZero() && item.UploadDate != "" {
				info.ReleaseDate = parser.ParseDate(item.UploadDate)
			}
			if info.Maker == "" {
				info.Maker = strings.TrimSpace(item.Brand.Name)
			}
			info.Genres = appendUnique(info.Genres, strings.TrimSpace(item.Category))
			for i, rawImage := range item.Image {
				image := absoluteFantiaURL(rawImage)
				if image == "" {
					continue
				}
				if i == 0 {
					info.CoverURL = image
					info.ThumbURL = image
				} else {
					info.PreviewImages = appendUnique(info.PreviewImages, image)
				}
			}
			if item.ThumbnailURL != "" {
				info.ThumbURL = absoluteFantiaURL(item.ThumbnailURL)
			}
		}
	})

	// GTM JSON-LD (Tags)
	c.OnHTML(`script.gtm-json`, func(e *colly.HTMLElement) {
		var data struct {
			Tag []string `json:"tag"`
		}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			gtmTags = data.Tag
		}
	})

	// Genres
	c.OnHTML(`div.fanclub-summary a.btn-category`, func(e *colly.HTMLElement) {
		info.Genres = append(info.Genres, strings.TrimSpace(e.Text))
	})

	// Genres (Fallback)
	c.OnHTML(`.product-category > div:nth-child(3) > a:nth-child(1)`, func(e *colly.HTMLElement) {
		info.Genres = append(info.Genres, strings.TrimSpace(e.Text))
	})

	// Release Date
	c.OnHTML(`div.product-date`, func(e *colly.HTMLElement) {
		info.ReleaseDate = parser.ParseDate(e.Text)
	})

	// Release Date (Fallback)
	c.OnHTML(`div.product-meta`, func(e *colly.HTMLElement) {
		if time.Time(info.ReleaseDate).IsZero() {
			if ss := regexp.MustCompile(`(\d{4}[-/]\d{1,2}[-/]\d{1,2})`).FindStringSubmatch(e.Text); len(ss) == 2 {
				info.ReleaseDate = parser.ParseDate(ss[1])
			}
		}
	})

	c.OnScraped(func(_ *colly.Response) {
		if len(gtmTags) > 0 {
			info.Genres = gtmTags
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func longerSummary(current, candidate string) string {
	candidate = strings.TrimSpace(candidate)
	if len(candidate) > len(current) {
		return candidate
	}
	return current
}

type schemaMovie struct {
	Type         string     `json:"@type"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	UploadDate   string     `json:"uploadDate"`
	ThumbnailURL string     `json:"thumbnailUrl"`
	Image        stringList `json:"image"`
	Category     string     `json:"category"`
	Brand        struct {
		Name string `json:"name"`
	} `json:"brand"`
}

type stringList []string

func (s *stringList) UnmarshalJSON(data []byte) error {
	var values []string
	if err := json.Unmarshal(data, &values); err == nil {
		*s = values
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = []string{value}
	return nil
}

func decodeSchemaMovies(raw string) []schemaMovie {
	data := bytes.TrimSpace([]byte(raw))
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		var movies []schemaMovie
		if json.Unmarshal(data, &movies) == nil {
			return movies
		}
		return nil
	}
	var movie schemaMovie
	if json.Unmarshal(data, &movie) == nil {
		return []schemaMovie{movie}
	}
	return nil
}

func productDescription(selection *goquery.Selection) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "h3" {
				for _, attr := range node.Attr {
					if attr.Key == "class" && strings.Contains(" "+attr.Val+" ", " content-title ") {
						return
					}
				}
			}
			if node.Data == "br" {
				b.WriteByte('\n')
				return
			}
		}
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if node.Type == html.ElementNode && (node.Data == "p" || node.Data == "div" || node.Data == "li") {
			b.WriteByte('\n')
		}
	}
	selection.Each(func(_ int, s *goquery.Selection) {
		for _, node := range s.Nodes {
			walk(node)
		}
	})
	lines := strings.Split(b.String(), "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			clean = append(clean, line)
		}
	}
	return strings.Join(clean, "\n")
}

func init() {
	provider.Register(Name, New)
	provider.Register(PostName, NewPost)
}
