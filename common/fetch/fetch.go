package fetch

import (
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var defaultFetcher = NewDefaultFetcher()

type Fetcher struct {
	httpClient *http.Client
}

func NewFetcher(c *http.Client) *Fetcher {
	return &Fetcher{
		httpClient: c,
	}
}

func NewDefaultFetcher() *Fetcher {
	return NewFetcher((&retryablehttp.Client{
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 3 * time.Second,
		RetryMax:     3,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
	}).StandardClient())
}

// Fetch fetches resources from url.
func (f *Fetcher) Fetch(url string, opts ...Option) (resp *http.Response, err error) {
	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return
	}
	// Apply options.
	for _, opt := range opts {
		opt(req)
	}
	// Make HTTP request.
	if resp, err = f.httpClient.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	return
}

func Fetch(url string, opts ...Option) (resp *http.Response, err error) {
	return defaultFetcher.Fetch(url, opts...)
}
