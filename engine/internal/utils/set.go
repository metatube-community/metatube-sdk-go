package utils

import (
	"github.com/iancoleman/orderedmap"

	"github.com/metatube-community/metatube-sdk-go/model"
)

type ActorSearchResultSet struct {
	o *orderedmap.OrderedMap
}

func NewActorSearchResultSet() *ActorSearchResultSet {
	return &ActorSearchResultSet{
		o: orderedmap.New(),
	}
}

func (sr *ActorSearchResultSet) Add(results ...*model.ActorSearchResult) {
	for _, result := range results {
		sr.o.Set(result.Provider+result.ID, result)
	}
}

func (sr *ActorSearchResultSet) Results() []*model.ActorSearchResult {
	results := make([]*model.ActorSearchResult, 0, len(sr.o.Keys()))
	for _, key := range sr.o.Keys() {
		v, _ := sr.o.Get(key)
		results = append(results, v.(*model.ActorSearchResult))
	}
	return results
}

type MovieSearchResultSet struct {
	o *orderedmap.OrderedMap
}

func NewMovieSearchResultSet() *MovieSearchResultSet {
	return &MovieSearchResultSet{
		o: orderedmap.New(),
	}
}

func (sr *MovieSearchResultSet) Add(results ...*model.MovieSearchResult) {
	for _, result := range results {
		sr.o.Set(result.Provider+result.ID, result)
	}
}

func (sr *MovieSearchResultSet) Results() []*model.MovieSearchResult {
	results := make([]*model.MovieSearchResult, 0, len(sr.o.Keys()))
	for _, key := range sr.o.Keys() {
		v, _ := sr.o.Get(key)
		results = append(results, v.(*model.MovieSearchResult))
	}
	return results
}
