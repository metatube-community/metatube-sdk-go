package graphql

import (
	"context"
	_ "embed"
	"net/http"

	"github.com/machinebox/graphql"
)

var (
	//go:embed query/SearchVideos.graphql
	searchVideosQuery string

	//go:embed query/FindOneVideo.graphql
	findOneVideoQuery string

	//go:embed query/SearchResultsTour.graphql
	searchResultsTourQuery string
)

const (
	skipSize   = 0
	searchSize = 10
)

type Client struct {
	gc  *graphql.Client
	ctx context.Context
}

func NewClient(url string) *Client {
	return &Client{
		gc:  graphql.NewClient(url, graphql.WithHTTPClient(http.DefaultClient)),
		ctx: context.Background(),
	}
}

func (c *Client) SearchVideos(query, site string) (*SearchVideosResponse, error) {
	req := graphql.NewRequest(searchVideosQuery)
	req.Var("query", query)
	req.Var("site", site)
	req.Var("first", searchSize)
	req.Var("skip", skipSize)

	var resp SearchVideosResponse
	if err := c.gc.Run(c.ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) SearchResultsTour(query, site string) (*SearchResultsTourResponse, error) {
	req := graphql.NewRequest(searchResultsTourQuery)
	req.Var("query", query)
	req.Var("site", site)
	req.Var("first", searchSize)
	req.Var("skip", skipSize)

	var resp SearchResultsTourResponse
	if err := c.gc.Run(c.ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) FindOneVideo(slug, site string) (*FindOneVideoResponse, error) {
	req := graphql.NewRequest(findOneVideoQuery)
	req.Var("slug", slug)
	req.Var("site", site)

	var resp FindOneVideoResponse
	if err := c.gc.Run(c.ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
