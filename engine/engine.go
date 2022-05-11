package engine

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

type Engine struct {
	movieProviders map[string]javtube.MovieProvider
	actorProviders map[string]javtube.ActorProvider
}

func New(timeout time.Duration) (engine *Engine) {
	engine = &Engine{
		movieProviders: make(map[string]javtube.MovieProvider),
		actorProviders: make(map[string]javtube.ActorProvider),
	}
	javtube.RangeMovieFactory(func(name string, factory javtube.MovieFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		engine.movieProviders[name] = provider
	})
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		engine.actorProviders[name] = factory()
	})
	return
}

func (e *Engine) searchMovie(provider javtube.MovieProvider, keyword string) ([]*model.MovieSearchResult, error) {
	if searcher, ok := provider.(javtube.MovieSearcher); ok {
		return searcher.SearchMovie(keyword)
	}
	info, err := e.getMovieInfoByID(provider, keyword)
	if err != nil {
		return nil, err
	}
	return []*model.MovieSearchResult{info.ToSearchResult()}, nil
}

func (e *Engine) SearchMovie(name string, keyword string) ([]*model.MovieSearchResult, error) {
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
func (e *Engine) SearchMovieAll(keyword string) ([]*model.MovieSearchResult, error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}

	type response struct {
		provider javtube.MovieProvider
		results  []*model.MovieSearchResult
		err      error
	}
	respCh := make(chan response)

	var wg sync.WaitGroup
	for _, provider := range e.movieProviders {
		wg.Add(1)
		// Async searching.
		go func(provider javtube.MovieProvider) {
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
		result   *model.MovieSearchResult
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
	var results []*model.MovieSearchResult
	for _, i := range items {
		results = append(results, i.result)
	}
	return results, nil
}

func (e *Engine) getMovieInfoByID(provider javtube.MovieProvider, id string) (info *model.MovieInfo, err error) {
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
