package solverr

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SkYNewZ/go-flaresolverr"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Client struct {
	client  flaresolverr.Client
	session uuid.UUID
}

func New(url string, timeout time.Duration, session uuid.UUID) *Client {
	return &Client{
		client:  flaresolverr.New(url, timeout, nil),
		session: session,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var (
		resp *flaresolverr.Response
		err  error
	)
	switch req.Method {
	case http.MethodGet:
		resp, err = c.client.Get(context.Background(), req.URL.String(), c.session)
	case http.MethodPost:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported method: %s", req.Method)
	}
	if err != nil {
		return nil, err
	}

	return &http.Response{
		Status:       http.StatusText(http.StatusOK),
		StatusCode:   http.StatusOK,
		Uncompressed: true,
		Request:      req,
		Body:         newReadCloserString(resp.Solution.Response),
	}, nil
}

func (c *Client) Get(rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return c.Do(&http.Request{Method: http.MethodGet, URL: u})
}

func (c *Client) StandardClient() *http.Client {
	return &http.Client{
		Transport: &RoundTripper{Client: c},
	}
}
