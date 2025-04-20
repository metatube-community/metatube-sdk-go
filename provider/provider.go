package provider

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/model"
)

type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// Priority returns the matching priority of the provider.
	Priority() float64

	// SetPriority sets the provider priority to the given value.
	SetPriority(v float64)

	// Language returns the primary language supported by the provider.
	Language() language.Tag

	// URL returns the base url of the provider.
	URL() *url.URL
}

type MovieSearcher interface {
	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.MovieSearchResult, error)

	// NormalizeMovieKeyword converts movie keyword to provider-friendly form.
	NormalizeMovieKeyword(Keyword string) string
}

type MovieReviewer interface {
	// GetMovieReviewsByID gets the user reviews of given movie id.
	GetMovieReviewsByID(id string) ([]*model.MovieReviewDetail, error)

	// GetMovieReviewsByURL gets the user reviews of given movie URL.
	GetMovieReviewsByURL(rawURL string) ([]*model.MovieReviewDetail, error)
}

type MovieProvider interface {
	// Provider should be implemented.
	Provider

	// NormalizeMovieID normalizes movie ID to conform to standard.
	NormalizeMovieID(id string) string

	// ParseMovieIDFromURL parses movie ID from given URL.
	ParseMovieIDFromURL(rawURL string) (string, error)

	// GetMovieInfoByID gets movie's info by id.
	GetMovieInfoByID(id string) (*model.MovieInfo, error)

	// GetMovieInfoByURL gets movie's info by url.
	GetMovieInfoByURL(url string) (*model.MovieInfo, error)
}

type ActorSearcher interface {
	// SearchActor searches matched actor/s.
	SearchActor(keyword string) ([]*model.ActorSearchResult, error)
}

type ActorProvider interface {
	// Provider should be implemented.
	Provider

	// NormalizeActorID normalizes actor ID to conform to standard.
	NormalizeActorID(id string) string

	// ParseActorIDFromURL parses actor ID from given URL.
	ParseActorIDFromURL(rawURL string) (string, error)

	// GetActorInfoByID gets actor's info by id.
	GetActorInfoByID(id string) (*model.ActorInfo, error)

	// GetActorInfoByURL gets actor's info by url.
	GetActorInfoByURL(url string) (*model.ActorInfo, error)
}

type Fetcher interface {
	// Fetch fetches media resources from url.
	Fetch(url string) (*http.Response, error)
}

type RequestTimeoutSetter interface {
	// SetRequestTimeout sets timeout for HTTP requests.
	SetRequestTimeout(timeout time.Duration)
}

type Config interface {
	Has(string) bool
	GetString(string) (string, error)
	GetBool(string) (bool, error)
	GetInt64(string) (int64, error)
	GetFloat64(string) (float64, error)
	GetDuration(string) (time.Duration, error)
}

type ConfigSetter interface {
	// SetConfig sets any additional configs for Provider.
	SetConfig(config Config) error
}
