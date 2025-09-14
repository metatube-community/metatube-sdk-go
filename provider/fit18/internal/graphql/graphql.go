package graphql

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/machinebox/graphql"
)

var (
	//go:embed query/Search.graphql
	searchQuery string

	//go:embed query/BatchFindAssetQuery.graphql
	batchFindAssetQuery string

	//go:embed query/FindVideo.graphql
	findVideoQuery string

	// debug enables or disables debug logging
	debug = os.Getenv("FIT18_DEBUG") == "true"
)

type Client struct {
	gc     *graphql.Client
	ctx    context.Context
	apiKey string
}

func NewClient(baseURL, apiKey string) *Client {
	httpClient := &http.Client{
		Transport: &headerTransport{
			Transport: http.DefaultTransport,
			Headers: map[string]string{
				"Content-type":                 "application/json",
				"argonath-api-key":             apiKey,
				"User-Agent":                   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0",
				"origin":                       baseURL,
				"referer":                      baseURL,
				"apollographql-client-name":    "fit18:site",
				"apollographql-client-version": "1.0",
				"accept":                       "*/*",
				"accept-language":              "en-US,en;q=0.6",
			},
		},
	}

	// Create the GraphQL client
	client := graphql.NewClient(baseURL, graphql.WithHTTPClient(httpClient))
	client.Log = func(s string) {
		if debug {
			log.Println(s)
		}
	}
	return &Client{
		gc:     client,
		ctx:    context.Background(),
		apiKey: apiKey,
	}
}

// headerTransport is custom transport that adds headers to each request and logs responses
type headerTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add headers to the request
	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}
	// Use the underlying transport to perform the request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Read and log the response body for debugging
	if resp != nil && resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			// If we can't read the body, just return the original response
			return resp, nil
		}

		// Restore the body for further processing
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return resp, nil
}

func (c *Client) Search(query string) (*SearchResponse, error) {
	req := graphql.NewRequest(searchQuery)
	req.Var("query", query)

	var resp SearchResponse
	if err := c.gc.Run(c.ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("error calling Search API: %w", err)
	}

	return &resp, nil
}

func (c *Client) BatchFindAsset(paths []string) (*BatchFindAssetResponse, error) {
	req := graphql.NewRequest(batchFindAssetQuery)
	req.Var("paths", paths)

	var resp BatchFindAssetResponse
	if err := c.gc.Run(c.ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("error calling BatchFindAsset API: %w", err)
	}

	return &resp, nil
}

func (c *Client) FindVideo(videoID string) (*FindVideoResponse, error) {
	req := graphql.NewRequest(findVideoQuery)
	req.Var("videoId", videoID)

	var resp FindVideoResponse
	if err := c.gc.Run(c.ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("error calling FindVideo API: %w", err)
	}

	return &resp, nil
}
