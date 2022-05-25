package provider

import (
	"net/http"
	"time"

	"github.com/javtube/javtube-sdk-go/model"
)

type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// URL returns the base url of the provider.
	URL() string

	// Priority returns the matching priority of the provider.
	Priority() int

	// NormalizeID normalizes ID to conform to standard.
	NormalizeID(id string) string
}

type MovieSearcher interface {
	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.MovieSearchResult, error)

	// TidyKeyword converts keyword to provider-friendly form.
	TidyKeyword(Keyword string) string
}

type MovieProvider interface {
	// Provider should be implemented.
	Provider

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
