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

type QueryOption struct {
	IsLoggedIn bool
	IsAmateur  bool
	IsAnime    bool
	IsAv       bool
	IsCinema   bool
	IsSP       bool
}

func NewClient() *Client {
	return &Client{
		c: graphql.NewClient(graphqlURL),
	}
}

func (c *Client) GetPPVContent(id string, opt QueryOption) (*PPVContent, error) {
	req := graphql.NewRequest(videoQuery)
	req.Var("id", id)
	req.Var("isLoggedIn", opt.IsLoggedIn)
	req.Var("isAmateur", opt.IsAmateur)
	req.Var("isAnime", opt.IsAnime)
	req.Var("isAv", opt.IsAv)
	req.Var("isCinema", opt.IsCinema)
	req.Var("isSP", opt.IsSP)

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
