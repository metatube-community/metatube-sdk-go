package scraper

import (
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/random"
)

type Option func(*Scraper)

func AllowURLRevisit() Option {
	return func(s *Scraper) {
		colly.AllowURLRevisit()(s.c)
	}
}

func WithCookies(url string, cookies []*http.Cookie) Option {
	return func(s *Scraper) {
		s.c.SetCookies(url, cookies)
	}
}

func WithDisableCookies() Option {
	return func(s *Scraper) {
		s.c.DisableCookies()
	}
}

func DetectCharset() Option {
	return func(s *Scraper) {
		colly.DetectCharset()(s.c)
	}
}

func IgnoreRobotsTxt() Option {
	return func(s *Scraper) {
		colly.IgnoreRobotsTxt()(s.c)
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(s *Scraper) {
		colly.Headers(headers)(s.c)
	}
}

func WithUserAgent(ua string) Option {
	return func(s *Scraper) {
		colly.UserAgent(ua)(s.c)
	}
}

func WithRandomUserAgent() Option {
	return WithUserAgent(random.UserAgent())
}
