package provider

import (
	"errors"

	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/util/random"
)

var (
	// UA is the default user agent for each collector.
	UA = random.UserAgent()
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupported   = errors.New("not supported")
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
