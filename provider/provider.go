package provider

import (
	"net/http"
	"net/url"
	"time"

	"github.com/metatube-community/metatube-sdk-go/model"
)

type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// Priority returns the matching priority of the provider.
	Priority() int

	// URL returns the base url of the provider.
	URL() *url.URL
}

type MovieSearcher interface {
	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.MovieSearchResult, error)

	// NormalizeKeyword converts keyword to provider-friendly form.
	NormalizeKeyword(Keyword string) string
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
