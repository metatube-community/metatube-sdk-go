package fetch

import (
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/javtube/javtube-sdk-go/common/random"
)

var DefaultFetcher = Default(nil)

type Config struct {
	// Set User-Agent Header.
	UserAgent string

	// Set Referer Header.
	Referer string

	// Use random User-Agent.
	RandomUserAgent bool
}

type Fetcher struct {
	client *http.Client
	config *Config
}

func New(c *http.Client, cfg *Config) *Fetcher {
	if cfg == nil /* init */ {
		cfg = new(Config)
	}
	if cfg.RandomUserAgent {
		// assign a random user-agent.
		cfg.UserAgent = random.UserAgent()
	}
	return &Fetcher{
		client: c,
		config: cfg,
	}
}

func Default(cfg *Config) *Fetcher {
	return New((&retryablehttp.Client{
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 3 * time.Second,
		RetryMax:     3,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
	}).StandardClient(), cfg)
}

// Fetch uses Get to fetch resources from url.
func (f *Fetcher) Fetch(url string) (resp *http.Response, err error) {
	return f.Get(url)
}

// Get gets resources from url with options.
func (f *Fetcher) Get(url string, opts ...Option) (resp *http.Response, err error) {
	return f.Request(http.MethodGet, url, opts...)
}

// Request requests resources with given method.
func (f *Fetcher) Request(method, url string, opts ...Option) (resp *http.Response, err error) {
	var req *http.Request
	if req, err = http.NewRequest(method, url, nil); err != nil {
		return
	}
	// compose options.
	var options []Option
	if ua := f.config.UserAgent; ua != "" {
		options = append(options, WithUserAgent(ua))
	}
	if referer := f.config.Referer; referer != "" {
		options = append(options, WithReferer(referer))
	}
	// apply options.
	for _, option := range append(options, opts...) {
		option.Apply(req)
	}
	// Make HTTP request.
	if resp, err = f.client.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	return
}

func Fetch(url string) (resp *http.Response, err error) {
	return DefaultFetcher.Fetch(url)
}

func Get(url string, opts ...Option) (resp *http.Response, err error) {
	return DefaultFetcher.Get(url, opts...)
}

func Request(method, url string, opts ...Option) (resp *http.Response, err error) {
	return DefaultFetcher.Request(method, url, opts...)
}

// Ignore warnings.
var (
	_ = Fetch
	_ = Get
	_ = Request
)
