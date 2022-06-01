package fetch

import (
	"net/http"

	"github.com/javtube/javtube-sdk-go/common/random"
)

type Option func(*http.Request)

func WithHeader(key, value string) Option {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func WithReferer(referer string) Option {
	return WithHeader("Referer", referer)
}

func WithUserAgent(ua string) Option {
	return WithHeader("User-Agent", ua)
}

func WithRandomUserAgent() Option {
	return WithUserAgent(random.UserAgent())
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
