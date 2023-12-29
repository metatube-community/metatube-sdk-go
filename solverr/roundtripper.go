package solverr

import (
	"net/http"
)

type RoundTripper struct {
	Client *Client
}

// RoundTrip satisfies the http.RoundTripper interface.
func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.Client.Do(req)
}
