package provider

import (
	"github.com/gocolly/colly/v2"
)

// Scraper implements basic Provider interface.
type Scraper struct {
	name     string
	priority int
	c        *colly.Collector
}

// NewScraper returns Provider implemented *Scraper.
func NewScraper(name string, priority int, c *colly.Collector) *Scraper {
	return &Scraper{
		name:     name,
		priority: priority,
		c:        c,
	}
}

func (s *Scraper) Name() string {
	return s.name
}

func (s *Scraper) Priority() int {
	return s.priority
}

func (s *Scraper) Collector() *colly.Collector {
	return s.c.Clone()
}
