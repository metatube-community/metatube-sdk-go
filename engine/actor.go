package engine

import (
	"fmt"
	"sort"
	"sync"

	"gorm.io/gorm/clause"

	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/gfriends"
)

func (e *Engine) searchActorFromDB(keyword string, provider javtube.Provider, all bool) (results []*model.ActorSearchResult, err error) {
	var infos []*model.ActorInfo
	if all {
		err = e.db.
			Where("name = ?", keyword).
			Find(&infos).Error
	} else {
		err = e.db.
			Where("provider = ? AND name = ?",
				provider.Name(), keyword).
			Find(&infos).Error
	}
	if err == nil {
		for _, info := range infos {
			if !info.Valid() {
				continue
			}
			results = append(results, info.ToSearchResult())
		}
	}
	return
}

func (e *Engine) searchActor(keyword string, provider javtube.Provider, fallback bool) ([]*model.ActorSearchResult, error) {
	innerSearch := func(keyword string) (results []*model.ActorSearchResult, err error) {
		if provider.Name() == gfriends.Name {
			return provider.(javtube.ActorSearcher).SearchActor(keyword)
		}
		if searcher, ok := provider.(javtube.ActorSearcher); ok {
			if fallback {
				defer func() {
					if innerResults, innerErr := e.searchActorFromDB(keyword, provider, false);
					// ignore DB query error.
					innerErr == nil && len(innerResults) > 0 {
						// overwrite error.
						err = nil
						// update results.
						asr := newActorSearchResults()
						asr.Add(results...)
						asr.Add(innerResults...)
						results = asr.Results()
					}
				}()
			}
			return searcher.SearchActor(keyword)
		}
		// All providers should implement ActorSearcher interface.
		return nil, javtube.ErrInfoNotFound
	}
	names := parser.ParseActorNames(keyword)
	if len(names) == 0 {
		return nil, javtube.ErrInvalidKeyword
	}
	var (
		results []*model.ActorSearchResult
		errors  []error
	)
	for _, name := range names {
		innerResults, innerErr := innerSearch(name)
		if innerErr != nil &&
			// ignore InfoNotFound error.
			innerErr != javtube.ErrInfoNotFound {
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
		return nil, javtube.ErrInfoNotFound
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

func (e *Engine) SearchActorAll(keyword string, fallback bool) (results []*model.ActorSearchResult, err error) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, provider := range e.actorProviders {
		wg.Add(1)
		go func(provider javtube.ActorProvider) {
			defer wg.Done()
			if innerResults, innerErr := e.searchActor(keyword, provider, false); innerErr == nil {
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

	if fallback {
		if innerResults, innerErr := e.searchActorFromDB(keyword, nil, true);
		// ignore DB query error.
		innerErr == nil && len(innerResults) > 0 {
			// overwrite error.
			err = nil
			// update results.
			asr := newActorSearchResults()
			asr.Add(results...)
			asr.Add(innerResults...)
			results = asr.Results()
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return e.MustGetActorProviderByName(results[i].Provider).Priority() >
			e.MustGetActorProviderByName(results[j].Provider).Priority()
	})
	return
}

func (e *Engine) getActorInfoFromDB(provider javtube.ActorProvider, id string) (*model.ActorInfo, error) {
	info := &model.ActorInfo{}
	err := e.db. // Exact match here.
			Where("provider = ?", provider.Name()).
			Where("id = ?", id).
			First(info).Error
	return info, err
}

func (e *Engine) getActorInfoWithCallback(provider javtube.ActorProvider, id string, lazy bool, callback func() (*model.ActorInfo, error)) (info *model.ActorInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.Valid()) {
			err = javtube.ErrIncompleteMetadata
		}
	}()
	if provider.Name() == gfriends.Name {
		return provider.GetActorInfoByID(id)
	}
	// Query DB first (by id).
	if lazy {
		if info, err = e.getActorInfoFromDB(provider, id); err == nil && info.Valid() {
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

func (e *Engine) getActorInfoByProviderID(provider javtube.ActorProvider, id string, lazy bool) (*model.ActorInfo, error) {
	if id = provider.NormalizeID(id); id == "" {
		return nil, javtube.ErrInvalidID
	}
	return e.getActorInfoWithCallback(provider, id, lazy, func() (*model.ActorInfo, error) {
		return provider.GetActorInfoByID(id)
	})
}

func (e *Engine) GetActorInfoByProviderID(name, id string, lazy bool) (*model.ActorInfo, error) {
	provider, err := e.GetActorProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoByProviderID(provider, id, lazy)
}

func (e *Engine) getActorInfoByProviderURL(provider javtube.ActorProvider, rawURL string, lazy bool) (*model.ActorInfo, error) {
	id, err := provider.ParseIDFromURL(rawURL)
	switch {
	case err != nil:
		return nil, err
	case id == "":
		return nil, javtube.ErrInvalidURL
	}
	return e.getActorInfoWithCallback(provider, id, lazy, func() (*model.ActorInfo, error) {
		return provider.GetActorInfoByURL(rawURL)
	})
}

func (e *Engine) GetActorInfoByURL(rawURL string, lazy bool) (*model.ActorInfo, error) {
	provider, err := e.GetActorProviderByURL(rawURL)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoByProviderURL(provider, rawURL, lazy)
}
