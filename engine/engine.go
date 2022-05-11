package engine

import (
	"fmt"
	"sort"
	"sync"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

type Engine struct {
	movieProviders map[string]javtube.Provider
	actorProviders map[string]javtube.ActorProvider
}

func New() (engine *Engine) {
	engine = &Engine{
		movieProviders: make(map[string]javtube.Provider),
		actorProviders: make(map[string]javtube.ActorProvider),
	}
	javtube.RangeFactory(func(name string, factory javtube.Factory) {
		engine.movieProviders[name] = factory()
	})
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		engine.actorProviders[name] = factory()
	})
	return
}

func (e *Engine) searchMovie(provider javtube.Provider, keyword string) (results []*model.SearchResult, err error) {
	// query DB first (by number)
	if searcher, ok := provider.(javtube.Searcher); ok {
		return searcher.SearchMovie(keyword)
	}
	var info *model.MovieInfo
	if info, err = e.getMovieInfoByID(provider, keyword); err == nil && info.Valid() {
		return []*model.SearchResult{info.ToSearchResult()}, nil
	}
	return
}

func (e *Engine) SearchMovie(name string, keyword string) (results []*model.SearchResult, err error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}
	provider, ok := e.movieProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return e.searchMovie(provider, keyword)
}

// SearchMovieAll searches the keyword from all providers.
func (e *Engine) SearchMovieAll(keyword string) ([]*model.SearchResult, error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}

	type response struct {
		provider javtube.Provider
		results  []*model.SearchResult
		err      error
	}
	respCh := make(chan response)

	var wg sync.WaitGroup
	for _, provider := range e.movieProviders {
		wg.Add(1)
		// Async searching.
		go func(provider javtube.Provider) {
			defer wg.Done()
			results, err := e.searchMovie(provider, keyword)
			respCh <- response{
				provider: provider,
				results:  results,
				err:      err,
			}
		}(provider)
	}
	go func() {
		wg.Wait()
		// notify when all searching tasks done.
		close(respCh)
	}()

	type item struct {
		priority float64
		result   *model.SearchResult
	}
	var items []item
	for resp := range respCh {
		if resp.err != nil {
			continue
		}
		for _, result := range resp.results {
			if !result.Valid() {
				continue
			}
			items = append(items, item{
				// calculate priority.
				priority: float64(resp.provider.Priority()) *
					number.Similarity(keyword, result.Number),
				result: result,
			})
		}
	}
	// sort items according to its priority.
	sort.SliceStable(items, func(i, j int) bool {
		// higher priority comes first.
		return items[i].priority > items[j].priority
	})
	// refine search results.
	var results []*model.SearchResult
	for _, i := range items {
		results = append(results, i.result)
	}
	return results, nil
}

func (e *Engine) getMovieInfoByID(provider javtube.Provider, id string) (info *model.MovieInfo, err error) {
	// query DB (by id)
	//
	// defer save
	defer func() {
		if err == nil && info.Valid() {
			// save to DB
		}
	}()
	return provider.GetMovieInfoByID(id)
}

func (e *Engine) GetMovieInfoByID(name, id string) (info *model.MovieInfo, err error) {
	// query DB (by id)
	//
	// defer save
	provider, ok := e.movieProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return e.getMovieInfoByID(provider, id)
}

func (e *Engine) SearchActor() {

}

func (e *Engine) GetActorInfo() {

}

func (e *Engine) GetImage() {

}
