package engine

import (
	"fmt"
	"sort"
	"sync"

	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/gfriends"
	"gorm.io/gorm/clause"
)

func (e *Engine) searchActor(keyword string, provider javtube.Provider, lazy bool) (results []*model.ActorSearchResult, err error) {
	if provider.Name() == gfriends.Name {
		return provider.(javtube.ActorSearcher).SearchActor(keyword)
	}
	if searcher, ok := provider.(javtube.ActorSearcher); ok {
		// Query DB first (by name or id).
		if info := new(model.ActorInfo); lazy {
			if result := e.db.
				Where("provider = ?", provider.Name()).
				Where(e.db.
					Where("name = ?", keyword).
					Or("id = ?", keyword)).
				First(info); result.Error == nil && info.Valid() /* must be valid */ {
				return []*model.ActorSearchResult{info.ToSearchResult()}, nil
			}
		}
		return searcher.SearchActor(keyword)
	}
	// All providers should implement ActorSearcher interface.
	return nil, javtube.ErrNotFound
}

func (e *Engine) SearchActor(keyword, name string, lazy bool) (results []*model.ActorSearchResult, err error) {
	provider, ok := e.actorProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return e.searchActor(keyword, provider, lazy)
}

func (e *Engine) SearchActorAll(keyword string) (results []*model.ActorSearchResult, err error) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, provider := range e.actorProviders {
		wg.Add(1)
		go func(provider javtube.ActorProvider) {
			defer wg.Done()
			if innerResults, innerErr := e.searchActor(keyword, provider, true); innerErr == nil {
				for _, result := range innerResults {
					if result.Valid() {
						mu.Lock()
						results = append(results, result)
						mu.Unlock()
					}
				}
			} // ignore error
		}(provider)
	}
	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return e.actorProviders[results[i].Provider].Priority() >
			e.actorProviders[results[j].Provider].Priority()
	})
	return
}

func (e *Engine) getActorInfoByID(id string, provider javtube.ActorProvider, lazy bool) (info *model.ActorInfo, err error) {
	if id = provider.NormalizeID(id); id == "" {
		return nil, javtube.ErrInvalidID
	}
	if provider.Name() == gfriends.Name {
		return provider.GetActorInfoByID(id)
	}
	// Query DB first (by id).
	if info = new(model.ActorInfo); lazy {
		if result := e.db.
			// Exact match here.
			Where("id = ?", id).
			Where("provider = ?", provider.Name()).
			First(info); result.Error == nil && info.Valid() {
			return
		}
	}
	// Delayed info auto-save.
	defer func() {
		if err == nil && info.Valid() {
			// Make sure we save the original info here.
			e.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(info) // ignore error
		}
	}()
	return provider.GetActorInfoByID(id)
}

func (e *Engine) GetActorInfoByID(id, name string, lazy bool) (info *model.ActorInfo, err error) {
	provider, ok := e.actorProviders[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return e.getActorInfoByID(id, provider, lazy)
}
