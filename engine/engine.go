package engine

import (
	"fmt"
	"sort"
	"sync"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"gorm.io/gorm"
)

type Engine struct {
	db             *gorm.DB
	movieProviders map[string]javtube.Provider
	actorProviders map[string]javtube.ActorProvider
}

func New() *Engine {
	var (
		movieProviders = make(map[string]javtube.Provider)
		actorProviders = make(map[string]javtube.ActorProvider)
	)
	javtube.RangeFactory(func(name string, factory javtube.Factory) {
		movieProviders[name] = factory()
	})
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		actorProviders[name] = factory()
	})
	return &Engine{
		movieProviders: movieProviders,
		actorProviders: actorProviders,
	}
}

func (e *Engine) searchMovie(provider javtube.Provider, keyword string) (results []*model.SearchResult, err error) {
	// query DB first (by number)
	if searcher, ok := provider.(javtube.Searcher); ok {
		return searcher.SearchMovie(keyword)
	}
	var info *model.MovieInfo
	if info, err = provider.GetMovieInfoByID(keyword); err == nil && info.Valid() {
		return []*model.SearchResult{info.ToSearchResult()}, nil
	}
	return
}

func (e *Engine) SearchMovie(keyword string) ([]*model.SearchResult, error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}

	type response struct {
		priority int
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
				priority: provider.Priority(),
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
				priority: float64(resp.priority) *
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
	//
	var results []*model.SearchResult
	for _, i := range items {
		results = append(results, i.result)
	}
	return results, nil
}

func (e *Engine) getMovieInfoFromDB(name, id string) (info *model.MovieInfo, err error) {
	return nil, err
}

func (e *Engine) GetMovieInfo(name, id string) (info *model.MovieInfo, err error) {
	// query DB (by id)
	//
	// defer save
	defer func() {
		if err == nil && info.Valid() {
			// save to DB
		}
	}()
	provider, ok := e.movieProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider.GetMovieInfoByID(id)
}

func (e *Engine) SearchActor() {

}

func (e *Engine) GetActorInfo() {

}

func (e *Engine) GetImage() {

}
