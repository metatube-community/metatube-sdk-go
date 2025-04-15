package engine

import (
	goerr "errors"
	"fmt"
	"sort"
	"sync"

	"golang.org/x/text/language"
	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/collections"
	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *Engine) searchActorFromDB(keyword string, provider mt.Provider) (results []*model.ActorSearchResult, err error) {
	var infos []*model.ActorInfo
	if err = e.db.
		Where("provider = ? AND name = ? COLLATE NOCASE",
			provider.Name(), keyword).
		Find(&infos).Error; err == nil {
		for _, info := range infos {
			if !info.Valid() {
				continue
			}
			results = append(results, info.ToSearchResult())
		}
	}
	return
}

func (e *Engine) searchActor(keyword string, provider mt.Provider, fallback bool) ([]*model.ActorSearchResult, error) {
	innerSearch := func(keyword string) (results []*model.ActorSearchResult, err error) {
		if searcher, ok := provider.(mt.ActorSearcher); ok {
			defer func() {
				if err != nil || len(results) == 0 {
					return // ignore error or empty.
				}
				const minSimilarity = 0.3
				ps := new(collections.WeightedSlice[float64, *model.ActorSearchResult])
				for _, result := range results {
					if similarity := comparer.Compare(result.Name, keyword); similarity >= minSimilarity {
						ps.Append(similarity, result)
					}
				}
				results = ps.SortFunc(sort.Stable).Underlying() // replace results.
			}()
			if fallback {
				defer func() {
					if innerResults, innerErr := e.searchActorFromDB(keyword, provider);
					// ignore DB query error.
					innerErr == nil && len(innerResults) > 0 {
						// overwrite error.
						err = nil
						// update results.
						asr := collections.NewOrderedSet(func(v *model.ActorSearchResult) string { return v.Provider + v.ID })
						// unlike movie searching, we want search results go first
						// than DB data here, so we add results later than DB results.
						asr.Add(innerResults...)
						asr.Add(results...)
						results = asr.Slice()
					}
				}()
			}
			return searcher.SearchActor(keyword)
		}
		// All providers should implement ActorSearcher interface.
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

func (e *Engine) SearchActorAll(keyword string, fallback bool) (results []*model.ActorSearchResult, err error) {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, provider := range e.actorProviders {
		wg.Add(1)
		go func(provider mt.ActorProvider) {
			defer wg.Done()
			if innerResults, innerErr := e.searchActor(keyword, provider, fallback); innerErr == nil {
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

func (e *Engine) getActorInfoFromDB(provider mt.ActorProvider, id string) (*model.ActorInfo, error) {
	info := &model.ActorInfo{}
	err := e.db. // Exact match here.
			Where("provider = ?", provider.Name()).
			Where("id = ? COLLATE NOCASE", id).
			First(info).Error
	return info, err
}

func (e *Engine) getActorInfoWithCallback(provider mt.ActorProvider, id string, lazy bool, callback func() (*model.ActorInfo, error)) (*model.ActorInfo, error) {
	// Query DB first (by id).
	if lazy {
		if info, err := e.getActorInfoFromDB(provider, id); err == nil && info.Valid() {
			// actor image injection.
			e.injectActorImages(provider.Language(), info)
			return info, nil
		}
	}

	info, err := callback()
	if err != nil {
		return nil, err
	}

	// metadata validation check.
	if info == nil {
		return nil, mt.ErrInfoNotFound
	} else if !info.Valid() {
		return nil, mt.ErrIncompleteMetadata
	}

	// save info to db.
	e.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(info) // ignore error

	// actor image injection.
	e.injectActorImages(provider.Language(), info)

	return info, nil
}

func (e *Engine) injectActorImages(lang language.Tag, info *model.ActorInfo) {
	imageProviders := e.GetActorImageProviderByLanguage(lang)
	for _, imageProvider := range imageProviders {
		if images, err := imageProvider.GetActorImagesByName(info.Name); err == nil && len(images) > 0 {
			info.Images = append(info.Images, images...)
		}
	}
}

func (e *Engine) getActorInfoByProviderID(provider mt.ActorProvider, id string, lazy bool) (*model.ActorInfo, error) {
	if id = provider.NormalizeActorID(id); id == "" {
		return nil, mt.ErrInvalidID
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

func (e *Engine) getActorInfoByProviderURL(provider mt.ActorProvider, rawURL string, lazy bool) (*model.ActorInfo, error) {
	id, err := provider.ParseActorIDFromURL(rawURL)
	switch {
	case err != nil:
		return nil, err
	case id == "":
		return nil, mt.ErrInvalidURL
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
