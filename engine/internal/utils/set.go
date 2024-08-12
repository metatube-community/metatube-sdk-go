package utils

import (
	"github.com/iancoleman/orderedmap"

	"github.com/metatube-community/metatube-sdk-go/model"
)

type SearchResult interface {
	*model.ActorSearchResult | *model.MovieSearchResult
}

type SearchResultSet[T SearchResult] struct {
	o *orderedmap.OrderedMap
}

func NewSearchResultSet[T SearchResult]() *SearchResultSet[T] {
	return &SearchResultSet[T]{
		o: orderedmap.New(),
	}
}

func (sr *SearchResultSet[T]) Add(results ...T) {
	for _, result := range results {
		switch t := any(result).(type) {
		case *model.ActorSearchResult:
			sr.o.Set(t.Provider+t.ID, result)
		case *model.MovieSearchResult:
			sr.o.Set(t.Provider+t.ID, result)
		}
	}
}

func (sr *SearchResultSet[T]) Results() []T {
	results := make([]T, 0, len(sr.o.Keys()))
	for _, key := range sr.o.Keys() {
		v, _ := sr.o.Get(key)
		results = append(results, v.(T))
	}
	return results
}
