package fetch

import (
	"net/http"

	"github.com/metatube-community/metatube-sdk-go/common/random"
)

// Context is used for each request.
type Context struct {
	req *http.Request
	Config
}

type Option func(*Context)

func (opt Option) apply(c *Context) { opt(c) }

func WithRaiseForStatus(v bool) Option {
	return func(c *Context) { c.RaiseForStatus = v }
}

func WithRequest(fn func(req *http.Request)) Option {
	return func(c *Context) { fn(c.req) }
}

func WithHeader(key, value string) Option {
	return WithRequest(func(req *http.Request) {
		req.Header.Set(key, value)
	})
}

func WithHeaders(headers map[string]string) Option {
	return WithRequest(func(req *http.Request) {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	})
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
	return WithRequest(func(req *http.Request) {
		req.SetBasicAuth(username, password)
	})
}

func WithQuery(key, value string) Option {
	return WithRequest(func(req *http.Request) {
		q := req.URL.Query()
		q.Set(key, value)
		req.URL.RawQuery = q.Encode()
	})
}

func WithQueryMap(query map[string]string) Option {
	return WithRequest(func(req *http.Request) {
		q := req.URL.Query()
		for key, value := range query {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	})
}

func WithQueryPairs(kv ...string) Option {
	return WithRequest(func(req *http.Request) {
		q := req.URL.Query()
		if len(kv)%2 != 0 {
			panic("invalid key-value pairs")
		}
		for i := 0; i < len(kv); i += 2 {
			q.Set(kv[i], kv[i+1])
		}
		req.URL.RawQuery = q.Encode()
	})
}
