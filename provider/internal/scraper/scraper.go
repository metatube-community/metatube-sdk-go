package scraper

import (
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"

	"github.com/metatube-community/metatube-sdk-go/provider"
)

var _ provider.Provider = (*Scraper)(nil)

// Scraper implements basic Provider interface.
type Scraper struct {
	name     string
	priority int
	baseURL  *url.URL
	c        *colly.Collector
}

// NewScraper returns Provider implemented *Scraper.
func NewScraper(name, baseURL string, priority int, opts ...Option) *Scraper {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	s := &Scraper{
		name:     name,
		priority: priority,
		baseURL:  u,
		c:        colly.NewCollector(),
	}
	for _, opt := range opts {
		// Apply options.
		if err := opt(s); err != nil {
			panic(err)
		}
	}
	return s
}

// NewDefaultScraper returns a *Scraper with default options enabled.
func NewDefaultScraper(name, baseURL string, priority int, opts ...Option) *Scraper {
	return NewScraper(name, baseURL, priority, append([]Option{
		WithAllowURLRevisit(),
		WithIgnoreRobotsTxt(),
		WithRandomUserAgent(),
	}, opts...)...)
}

func (s *Scraper) Name() string { return s.name }

func (s *Scraper) URL() *url.URL { return s.baseURL }

func (s *Scraper) Priority() int { return s.priority }

func (s *Scraper) NormalizeID(id string) string { return id /* AS IS */ }

func (s *Scraper) ParseIDFromURL(string) (string, error) { panic("unimplemented") }

// ClonedCollector returns cloned internal collector.
func (s *Scraper) ClonedCollector() *colly.Collector { return s.c.Clone() }

// SetRequestTimeout sets timeout for HTTP requests.
func (s *Scraper) SetRequestTimeout(timeout time.Duration) { s.c.SetRequestTimeout(timeout) }
