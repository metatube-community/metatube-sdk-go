package graphql

import (
	"context"
	_ "embed"
	"errors"
	"reflect"

	"github.com/machinebox/graphql"
)

const (
	videoURL   = "https://video.dmm.co.jp/"
	graphqlURL = "https://api.video.dmm.co.jp/graphql"
)

var (
	//go:embed query/ContentPageData.graphql
	contentPageDataQuery string

	//go:embed query/UserReviews.graphql
	userReviewsQuery string
)

var ErrNullResponse = errors.New("response is null")

type ClientOption = graphql.ClientOption

var (
	WithHTTPClient   = graphql.WithHTTPClient
	UseMultipartForm = graphql.UseMultipartForm
)

type Client struct {
	gc *graphql.Client
}

func NewClient(opts ...ClientOption) *Client {
	return &Client{
		gc: graphql.NewClient(graphqlURL, opts...),
	}
}

func (c *Client) GetContentPageData(id string, opts ContentPageDataQueryOptions) (*ContentPageDataResponse, error) {
	req := graphql.NewRequest(contentPageDataQuery)
	req.Var("id", id)
	req.Var("isLoggedIn", opts.IsLoggedIn)
	req.Var("isAmateur", opts.IsAmateur)
	req.Var("isAnime", opts.IsAnime)
	req.Var("isAv", opts.IsAv)
	req.Var("isCinema", opts.IsCinema)
	req.Var("isSP", opts.IsSP)
	req.Var("shouldFetchRelatedTags", true)

	req.Header.Set("Referer", videoURL)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Fanza-Device", "BROWSER")
	req.Header.Set("User-Agent", "") // skip

	var resp ContentPageDataResponse
	if err := c.gc.Run(context.Background(), req, &resp); err != nil {
		return nil, err
	}

	if reflect.DeepEqual(resp, ContentPageDataResponse{Typename: resp.Typename}) {
		return nil, ErrNullResponse
	}

	return &resp, nil
}

func (c *Client) GetUserReviews(id string, offset ...int) (*UserReviewsResponse, error) {
	req := graphql.NewRequest(userReviewsQuery)
	req.Var("id", id)
	req.Var("sort", "HELPFUL_COUNT_DESC")
	req.Var("offset", 0) // default offset
	if len(offset) > 0 {
		req.Var("offset", offset[0])
	}

	req.Header.Set("Referer", videoURL)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Fanza-Device", "BROWSER")
	req.Header.Set("User-Agent", "") // skip

	var resp UserReviewsResponse
	if err := c.gc.Run(context.Background(), req, &resp); err != nil {
		return nil, err
	}

	if reflect.DeepEqual(resp, UserReviewsResponse{}) {
		return nil, ErrNullResponse
	}

	return &resp, nil
}
