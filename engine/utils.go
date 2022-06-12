package engine

import (
	"github.com/iancoleman/orderedmap"

	"github.com/javtube/javtube-sdk-go/model"
)

type actorSearchResults struct {
	o *orderedmap.OrderedMap
}

func newActorSearchResults() *actorSearchResults {
	return &actorSearchResults{
		o: orderedmap.New(),
	}
}

func (sr *actorSearchResults) Add(results ...*model.ActorSearchResult) {
	for _, result := range results {
		sr.o.Set(result.Provider+result.ID, result)
	}
}

func (sr *actorSearchResults) Results() []*model.ActorSearchResult {
	results := make([]*model.ActorSearchResult, 0, len(sr.o.Keys()))
	for _, key := range sr.o.Keys() {
		v, _ := sr.o.Get(key)
		results = append(results, v.(*model.ActorSearchResult))
	}
	return results
}

type movieSearchResults struct {
	o *orderedmap.OrderedMap
}

func newMovieSearchResults() *movieSearchResults {
	return &movieSearchResults{
		o: orderedmap.New(),
	}
}

func (sr *movieSearchResults) Add(results ...*model.MovieSearchResult) {
	for _, result := range results {
		sr.o.Set(result.Provider+result.ID, result)
	}
}

func (sr *movieSearchResults) Results() []*model.MovieSearchResult {
	results := make([]*model.MovieSearchResult, 0, len(sr.o.Keys()))
	for _, key := range sr.o.Keys() {
		v, _ := sr.o.Get(key)
		results = append(results, v.(*model.MovieSearchResult))
	}
	return results
}
