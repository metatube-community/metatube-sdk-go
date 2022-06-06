package fetch

import (
	"net/http"

	"github.com/javtube/javtube-sdk-go/common/random"
)

type Option func(*http.Request)

func (opt Option) apply(req *http.Request) {
	opt(req)
}

func WithHeader(key, value string) Option {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(req *http.Request) {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
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

func WithAuthorization(token string) Option {
	return WithHeader("Authorization", "Bearer "+token)
}

func WithBasicAuth(username, password string) Option {
	return func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
}

func WithQuery(kv ...string) Option {
	return func(req *http.Request) {
		q := req.URL.Query()
		if len(kv)%2 != 0 {
			panic("invalid key-value pairs")
		}
		for i := 0; i < len(kv); i += 2 {
			q.Set(kv[i], kv[i+1])
		}
		req.URL.RawQuery = q.Encode()
	}
}

func WithQueryMap(query map[string]string) Option {
	return func(req *http.Request) {
		q := req.URL.Query()
		for key, value := range query {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
}

func WithHookFunc(fn func(req *http.Request)) Option {
	return func(req *http.Request) {
		fn(req)
	}
}
