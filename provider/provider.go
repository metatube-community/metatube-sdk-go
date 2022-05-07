package provider

import (
	"io"

	"github.com/javtube/javtube-sdk-go/model"
)

type Downloader interface {
	// Download downloads media resources from link.
	Download(link string) (io.ReadCloser, error)
}

type Searcher interface {
	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.SearchResult, error)
}

type ActorSearcher interface {
	// SearchActor searches matched actor/s.
	SearchActor(keyword string) ([]*model.ActorSearchResult, error)
}

type Provider interface {
	// GetMovieInfoByID gets movie's info by id.
	GetMovieInfoByID(id string) (*model.MovieInfo, error)

	// GetMovieInfoByLink gets movie's info by link.
	GetMovieInfoByLink(link string) (*model.MovieInfo, error)
}

type ActorProvider interface {
	// GetActorInfoByID gets actor's info by id.
	GetActorInfoByID(id string) (*model.ActorInfo, error)

	// GetActorInfoByLink gets actor's info by link.
	GetActorInfoByLink(link string) (*model.ActorInfo, error)
}
