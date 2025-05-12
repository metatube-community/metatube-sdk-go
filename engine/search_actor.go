package engine

import (
	goerr "errors"
	"fmt"
	goslices "slices"
	"sort"

	"github.com/metatube-community/metatube-sdk-go/collection/sets"
	"github.com/metatube-community/metatube-sdk-go/collection/slices"
	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/common/parallel"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/engine/dbengine"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/gfriends"
)

func (e *Engine) searchActor(keyword string, provider mt.Provider, fallback bool) ([]*model.ActorSearchResult, error) {
	innerSearch := func(keyword string) (results []*model.ActorSearchResult, err error) {
		if provider.Name() == gfriends.Name {
			return provider.(mt.ActorSearcher).SearchActor(keyword)
		}
		if searcher, ok := provider.(mt.ActorSearcher); ok {
			defer func() {
				if err != nil || len(results) == 0 {
					return // ignore error or empty.
				}
				const minSimilarity = 0.3
				ps := new(slices.WeightedSlice[*model.ActorSearchResult, float64])
				for _, result := range results {
					if similarity := comparer.Compare(result.Name, keyword); similarity >= minSimilarity {
						ps.Append(result, similarity)
					}
				}
				results = ps.SortFunc(sort.Stable).Slice() // replace results.
			}()
			if fallback {
				defer func() {
					if innerResults, innerErr := e.db.SearchActor(
						keyword,
						dbengine.ActorSearchOptions{
							Provider: provider.Name(),
						},
					); innerErr == nil /* ignore DB query error */ && len(innerResults) > 0 {
						// overwrite error.
						err = nil
						// update results.
						asr := sets.NewOrderedSetWithHash(func(v *model.ActorSearchResult) string { return v.Provider + v.ID })
						// unlike movie searching, we want search results go first
						// than DB data here, so we add results later than DB results.
						asr.Add(innerResults...)
						asr.Add(results...)
						results = asr.AsSlice()
					}
				}()
			}
			return searcher.SearchActor(keyword)
		}
		// All providers should implement the ActorSearcher interface.
		return nil, mt.ErrInfoNotFound
	}
	names := parser.ParseActorNames(keyword)
	if len(names) == 0 {
		return nil, mt.ErrInvalidKeyword
	}
	var (
		results []*model.ActorSearchResult
		errors  []error
	)
	for _, name := range names {
		innerResults, innerErr := innerSearch(name)
		if innerErr != nil &&
			// ignore InfoNotFound error.
			!goerr.Is(innerErr, mt.ErrInfoNotFound) {
			// add error to chain and handle it later.
			errors = append(errors, innerErr)
			continue
		}
		results = append(results, innerResults...)
	}
	if len(results) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("search errors: %v", errors)
		}
		return nil, mt.ErrInfoNotFound
	}
	return results, nil
}

func (e *Engine) SearchActor(keyword, name string, fallback bool) ([]*model.ActorSearchResult, error) {
	provider, err := e.GetActorProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.searchActor(keyword, provider, fallback)
}

func (e *Engine) SearchActorAll(keyword string, fallback bool) ([]*model.ActorSearchResult, error) {
	searchActor := func(provider mt.ActorProvider) []*model.ActorSearchResult {
		results := make([]*model.ActorSearchResult, 0)
		if innerResults, innerErr := e.searchActor(keyword, provider, fallback); innerErr == nil {
			for _, result := range innerResults {
				if result.IsValid() /* validation check */ {
					results = append(results, result)
				}
			}
		} // ignore error
		return results
	}
	results := slices.Flatten(parallel.Parallel(
		searchActor,
		goslices.Collect(e.actorProviders.Values())...,
	))

	sort.SliceStable(results, func(i, j int) bool {
		return e.MustGetActorProviderByName(results[i].Provider).Priority() >
			e.MustGetActorProviderByName(results[j].Provider).Priority()
	})
	return results, nil
}
