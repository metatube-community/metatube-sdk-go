package fetch

import (
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var defaultClient = (&retryablehttp.Client{
	RetryWaitMin: 1 * time.Second,
	RetryWaitMax: 3 * time.Second,
	RetryMax:     3,
	CheckRetry:   retryablehttp.DefaultRetryPolicy,
	Backoff:      retryablehttp.DefaultBackoff,
}).StandardClient()

type Option func(*http.Request)

func WithReferer(referer string) Option {
	return func(req *http.Request) {
		req.Header.Set("Referer", referer)
	}
}

func WithHeader(key, value string) Option {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func WithQuery(query map[string]string) Option {
	return func(req *http.Request) {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
}

// Fetch fetches resources from url.
func Fetch(u string, opts ...Option) (resp *http.Response, err error) {
	var (
		req *http.Request
	)
	if req, err = http.NewRequest(http.MethodGet, u, nil); err != nil {
		return
	}
	// Apply options.
	for _, opt := range opts {
		opt(req)
	}
	// Make HTTP request.
	if resp, err = defaultClient.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	return
}
