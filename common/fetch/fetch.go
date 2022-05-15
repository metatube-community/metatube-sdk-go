package fetch

import (
	"errors"
	"net/http"
)

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
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	return
}
