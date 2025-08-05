package graphql

import (
	"context"
	_ "embed"

	"github.com/machinebox/graphql"
)

const (
	videoURL   = "https://video.dmm.co.jp/"
	graphqlURL = "https://api.video.dmm.co.jp/graphql"
)

//go:embed video.graphql
var videoQuery string

type Client struct {
	c *graphql.Client
}

func NewClient() *Client {
	return &Client{
		c: graphql.NewClient(graphqlURL),
	}
}

func (c *Client) GetPPVContent(id string, opts QueryOptions) (*PPVContent, error) {
	req := graphql.NewRequest(videoQuery)
	req.Var("id", id)
	req.Var("isLoggedIn", opts.IsLoggedIn)
	req.Var("isAmateur", opts.IsAmateur)
	req.Var("isAnime", opts.IsAnime)
	req.Var("isAv", opts.IsAv)
	req.Var("isCinema", opts.IsCinema)
	req.Var("isSP", opts.IsSP)

	req.Header.Set("Referer", videoURL)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Fanza-Device", "BROWSER")
	req.Header.Set("User-Agent", "") // skip

	var resp ResponseWrapper
	if err := c.c.Run(context.Background(), req, &resp); err != nil {
		return nil, err
	}

	return &resp.PPVContent, nil
}
