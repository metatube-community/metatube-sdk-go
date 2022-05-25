package engine

import (
	"sort"
	"sync"

	"gorm.io/gorm/clause"

	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/gfriends"
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

func (e *Engine) SearchActor(keyword, name string, lazy bool) ([]*model.ActorSearchResult, error) {
	provider, err := e.GetActorProviderByName(name)
	if err != nil {
		return nil, err
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
					if result.Valid() /* validation check */ {
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
		return e.MustGetActorProviderByName(results[i].Provider).Priority() >
			e.MustGetActorProviderByName(results[j].Provider).Priority()
	})
	return
}

func (e *Engine) getActorInfoFromDB(id string, provider javtube.ActorProvider) (*model.ActorInfo, error) {
	info := &model.ActorInfo{}
	err := e.db. // Exact match here.
			Where("id = ?", id).
			Where("provider = ?", provider.Name()).
			First(info).Error
	return info, err
}

func (e *Engine) getActorInfoWithCallback(id string, provider javtube.ActorProvider, lazy bool, callback func() (*model.ActorInfo, error)) (info *model.ActorInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.Valid()) {
			err = javtube.ErrInvalidMetadata
		}
	}()
	if id = provider.NormalizeID(id); id == "" {
		return nil, javtube.ErrInvalidID
	}
	if provider.Name() == gfriends.Name {
		return provider.GetActorInfoByID(id)
	}
	// Query DB first (by id).
	if lazy {
		if info, err = e.getActorInfoFromDB(id, provider); err == nil && info.Valid() {
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
	return callback()
}

func (e *Engine) getActorInfoByID(id string, provider javtube.ActorProvider, lazy bool) (*model.ActorInfo, error) {
	return e.getActorInfoWithCallback(id, provider, lazy,
		func() (*model.ActorInfo, error) {
			return provider.GetActorInfoByID(id)
		})
}

func (e *Engine) GetActorInfoByID(id, name string, lazy bool) (*model.ActorInfo, error) {
	provider, err := e.GetActorProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoByID(id, provider, lazy)
}

func (e *Engine) getActorInfoByURL(rawURL string, provider javtube.ActorProvider, lazy bool) (*model.ActorInfo, error) {
	id, err := provider.ParseIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoWithCallback(id, provider, lazy,
		func() (*model.ActorInfo, error) {
			return provider.GetActorInfoByURL(rawURL)
		})
}

func (e *Engine) GetActorInfoByURL(rawURL string, lazy bool) (*model.ActorInfo, error) {
	provider, err := e.GetActorProviderByURL(rawURL)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoByURL(rawURL, provider, lazy)
}
