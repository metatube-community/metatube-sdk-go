package scraper

import (
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/provider"
)

var _ provider.Provider = (*Scraper)(nil)

// Scraper implements basic Provider interface.
type Scraper struct {
	name     string
	priority int
	c        *colly.Collector
}

// NewScraper returns Provider implemented *Scraper.
func NewScraper(name string, priority int, opts ...colly.CollectorOption) *Scraper {
	c := colly.NewCollector(
		opts...)
	{ // default settings
		c.AllowURLRevisit = true
		c.IgnoreRobotsTxt = true
		c.UserAgent = random.UserAgent()
	}
	return &Scraper{
		name:     name,
		priority: priority,
		c:        c,
	}
}

func (s *Scraper) Name() string { return s.name }

func (s *Scraper) Priority() int { return s.priority }

func (s *Scraper) NormalizeID(id string) string { return id /* AS IS */ }

// Collector returns original/internal collector.
func (s *Scraper) Collector() *colly.Collector { return s.c }

// ClonedCollector returns cloned internal collector.
func (s *Scraper) ClonedCollector() *colly.Collector { return s.c.Clone() }

// SetRequestTimeout sets timeout for HTTP requests.
func (s *Scraper) SetRequestTimeout(timeout time.Duration) { s.c.SetRequestTimeout(timeout) }
