package provider

import (
	"github.com/javtube/javtube-sdk-go/model"
)

type Provider interface {
	// Name returns the provider's name.
	Name() string

	// GetMovieInfoByID gets movie's info by id.
	GetMovieInfoByID(id string) (*model.MovieInfo, error)

	// GetMovieInfoByLink gets movie's info by link.
	GetMovieInfoByLink(link string) (*model.MovieInfo, error)

	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.SearchResult, error)
}

type ActorProvider interface {
	// Name returns the provider's name.
	Name() string

	// GetActorInfoByID gets actor's info by id.
	GetActorInfoByID(id string) (*model.ActorInfo, error)

	// GetActorInfoByLink gets actor's info by link.
	GetActorInfoByLink(link string) (*model.ActorInfo, error)

	// SearchActor searches matched actor/s.
	SearchActor(keyword string) ([]*model.ActorSearchResult, error)
}
