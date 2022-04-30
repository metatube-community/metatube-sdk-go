package provider

import (
	"github.com/javtube/javtube/model"
)

type Provider interface {
	// GetMovieInfo gets movie's info by id.
	GetMovieInfo(id string) (*model.MovieInfo, error)

	// SearchMovie searches matched movies.
	SearchMovie(keyword string) ([]*model.SearchResult, error)
}
