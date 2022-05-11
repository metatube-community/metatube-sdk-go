package provider

import (
	"io"

	"github.com/javtube/javtube-sdk-go/model"
)

type Searcher interface {
	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.SearchResult, error)
}

type Provider interface {
	// Name returns name of the provider.
	Name() string

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
	// Name returns name of the provider.
	Name() string

	// GetActorInfoByID gets actor's info by id.
	GetActorInfoByID(id string) (*model.ActorInfo, error)

	// GetActorInfoByURL gets actor's info by url.
	GetActorInfoByURL(url string) (*model.ActorInfo, error)
}

type Downloader interface {
	// Download downloads media resources from url.
	Download(url string) (io.ReadCloser, error)
}
