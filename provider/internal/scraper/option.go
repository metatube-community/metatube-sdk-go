package scraper

import (
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"

	"github.com/metatube-community/metatube-sdk-go/common/random"
)

type Option func(*Scraper) error

func WithAllowURLRevisit() Option {
	return func(s *Scraper) error {
		colly.AllowURLRevisit()(s.c)
		return nil
	}
}

func WithLogDebugger() Option {
	return func(s *Scraper) error {
		colly.Debugger(&debug.LogDebugger{})
		return nil
	}
}

func WithDetectCharset() Option {
	return func(s *Scraper) error {
		colly.DetectCharset()(s.c)
		return nil
	}
}

func WithIgnoreRobotsTxt() Option {
	return func(s *Scraper) error {
		colly.IgnoreRobotsTxt()(s.c)
		return nil
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(s *Scraper) error {
		colly.Headers(headers)(s.c)
		return nil
	}
}

func WithUserAgent(ua string) Option {
	return func(s *Scraper) error {
		colly.UserAgent(ua)(s.c)
		return nil
	}
}

func WithRandomUserAgent() Option {
	return WithUserAgent(random.UserAgent())
}

func WithCookies(url string, cookies []*http.Cookie) Option {
	return func(s *Scraper) error {
		return s.c.SetCookies(url, cookies)
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(s *Scraper) error {
		s.c.SetRequestTimeout(timeout)
		return nil
	}
}

func WithDisableCookies() Option {
	return func(s *Scraper) error {
		s.c.DisableCookies()
		return nil
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(s *Scraper) error {
		s.c.WithTransport(transport)
		return nil
	}
}
