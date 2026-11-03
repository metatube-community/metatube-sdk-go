package javdb

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2/fc2util"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*JavDB)(nil)
	_ provider.Fetcher       = (*JavDB)(nil)

	scorePattern = regexp.MustCompile(`\d+(?:\.\d+)?`)
	datePattern  = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
)

const (
	Name     = "JavDB"
	Priority = 1000 - 10
	baseURL  = "https://javdb.com/"
)

// JavDB provides basic FC2 metadata from public search results. Detail pages
// may require login, so lookups use the FC2 number rather than a detail-page ID.
type JavDB struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *JavDB {
	return &JavDB{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese,
			scraper.WithHeaders(map[string]string{"Referer": baseURL})),
	}
}

func (jav *JavDB) NormalizeMovieID(id string) string {
	return fc2util.ParseNumber(strings.TrimSpace(id))
}

func (jav *JavDB) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") ||
		(u.Hostname() != "javdb.com" && u.Hostname() != "www.javdb.com") ||
		u.Path != "/search" {
		return "", provider.ErrInvalidURL
	}
	if id := jav.NormalizeMovieID(u.Query().Get("q")); id != "" {
		return id, nil
	}
	return "", provider.ErrInvalidURL
}

func (jav *JavDB) GetMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	id, err := jav.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	return jav.GetMovieInfoByID(id)
}

func (jav *JavDB) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	id = jav.NormalizeMovieID(id)
	if id == "" {
		return nil, provider.ErrInvalidID
	}

	q := url.Values{"q": {"FC2-" + id}, "f": {"all"}}
	homepage := jav.URL().ResolveReference(&url.URL{Path: "/search", RawQuery: q.Encode()}).String()
	var info *model.MovieInfo
	c := jav.ClonedCollector()
	c.OnHTML(".movie-list .item", func(e *colly.HTMLElement) {
		// Search pages also contain related items. Never use the first result
		// without checking its number, including its FC2 prefix.
		number := strings.TrimSpace(e.ChildText(".video-title strong"))
		if !strings.HasPrefix(strings.ToUpper(number), "FC2") ||
			jav.NormalizeMovieID(number) != id || info != nil {
			return
		}
		title := strings.TrimSpace(e.ChildAttr("a.box", "title"))
		if title == "" {
			title = strings.TrimSpace(strings.TrimPrefix(
				strings.TrimSpace(e.ChildText(".video-title")), number))
		}
		cover := e.ChildAttr(".cover img", "src")
		if cover == "" {
			cover = e.ChildAttr(".cover img", "data-src")
		}
		if cover != "" {
			cover = e.Request.AbsoluteURL(cover)
		}
		info = &model.MovieInfo{
			ID:            id,
			Number:        "FC2-" + id,
			Title:         title,
			Provider:      jav.Name(),
			Homepage:      homepage,
			ThumbURL:      cover,
			CoverURL:      cover,
			Score:         parser.ParseScore(scorePattern.FindString(e.ChildText(".score .value"))),
			ReleaseDate:   parser.ParseDate(datePattern.FindString(e.ChildText(".meta"))),
			Actors:        []string{},
			Genres:        []string{},
			PreviewImages: []string{},
		}
	})
	if err := c.Visit(homepage); err != nil {
		return nil, err
	}
	if info == nil {
		return nil, provider.ErrInfoNotFound
	}
	if !info.IsValid() {
		return nil, provider.ErrIncompleteMetadata
	}
	return info, nil
}

func init() {
	provider.Register(Name, New)
}
