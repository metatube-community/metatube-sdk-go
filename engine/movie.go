package engine

import (
	"sync"

	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/priority"
	"github.com/metatube-community/metatube-sdk-go/engine/internal/utils"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *Engine) searchMovieFromDB(keyword string, provider mt.MovieProvider, all bool) (results []*model.MovieSearchResult, err error) {
	var infos []*model.MovieInfo
	tx := e.db.
		// Note: keyword might be an ID or just a regular number, so we should
		// query both of them for best match. Also, case should not mater.
		Where("number = ? COLLATE NOCASE", keyword).
		Or("id = ? COLLATE NOCASE", keyword)
	if all {
		err = tx.Find(&infos).Error
	} else {
		err = e.db.
			Where("provider = ?", provider.Name()).
			Where(tx).
			Find(&infos).Error
	}
	if err == nil {
		for _, info := range infos {
			if !info.Valid() {
				// normally it is valid, but just in case.
				continue
			}
			results = append(results, info.ToSearchResult())
		}
	}
	return
}

func (e *Engine) searchMovie(keyword string, provider mt.MovieProvider, fallback bool) (results []*model.MovieSearchResult, err error) {
	// Regular keyword searching.
	if searcher, ok := provider.(mt.MovieSearcher); ok {
		if keyword = searcher.NormalizeKeyword(keyword); keyword == "" {
			return nil, mt.ErrInvalidKeyword
		}
		if fallback {
			defer func() {
				if innerResults, innerErr := e.searchMovieFromDB(keyword, provider, false);
				// ignore DB query error.
				innerErr == nil && len(innerResults) > 0 {
					// overwrite error.
					err = nil
					// update results.
					msr := utils.NewMovieSearchResultSet()
					msr.Add(results...)
					msr.Add(innerResults...)
					results = msr.Results()
				}
			}()
		}
		return searcher.SearchMovie(keyword)
	}
	// Fallback to movie info querying.
	info, err := e.getMovieInfoByProviderID(provider, keyword, true)
	if err != nil {
		return nil, err
	}
	return []*model.MovieSearchResult{info.ToSearchResult()}, nil
}

func (e *Engine) SearchMovie(keyword, name string, fallback bool) ([]*model.MovieSearchResult, error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, mt.ErrInvalidKeyword
	}
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.searchMovie(keyword, provider, fallback)
}

func (e *Engine) searchMovieAll(keyword string) (results []*model.MovieSearchResult, err error) {
	type response struct {
		Results []*model.MovieSearchResult
		Error   error
	}
	respCh := make(chan response)

	var wg sync.WaitGroup
	for _, provider := range e.movieProviders {
		wg.Add(1)
		// Async searching.
		go func(provider mt.MovieProvider) {
			defer wg.Done()
			innerResults, innerErr := e.searchMovie(keyword, provider, false)
			respCh <- response{
				Results: innerResults,
				Error:   innerErr,
			}
		}(provider)
	}
	go func() {
		wg.Wait()
		// notify when all searching tasks done.
		close(respCh)
	}()

	// response channel.
	for resp := range respCh {
		if resp.Error != nil {
			continue
		}
		results = append(results, resp.Results...)
	}
	return
}

// SearchMovieAll searches the keyword from all providers.
func (e *Engine) SearchMovieAll(keyword string, fallback bool) (results []*model.MovieSearchResult, err error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, mt.ErrInvalidKeyword
	}

	defer func() {
		if err != nil {
			return
		}
		if len(results) == 0 {
			err = mt.ErrInfoNotFound
			return
		}
		// remove duplicate results, if any.
		msr := utils.NewMovieSearchResultSet()
		msr.Add(results...)
		results = msr.Results()
		// post-processing
		ps := new(priority.Slice[float64, *model.MovieSearchResult])
		for _, result := range results {
			if !result.Valid() /* validation check */ {
				continue
			}
			ps.Append(comparer.Compare(keyword, result.Number)*float64(e.MustGetMovieProviderByName(result.Provider).Priority()), result)
		}
		// sort according to priority.
		results = ps.Stable().Underlying()
	}()

	if fallback /* query database for missing results  */ {
		defer func() {
			if innerResults, innerErr := e.searchMovieFromDB(keyword, nil, true);
			// ignore DB query error.
			innerErr == nil && len(innerResults) > 0 {
				// overwrite error.
				err = nil
				// append results.
				results = append(results, innerResults...)
			}
		}()
	}

	results, err = e.searchMovieAll(keyword)
	return
}

func (e *Engine) getMovieInfoFromDB(provider mt.MovieProvider, id string) (*model.MovieInfo, error) {
	info := &model.MovieInfo{}
	err := e.db. // Exact match here.
			Where("provider = ?", provider.Name()).
			Where("id = ? COLLATE NOCASE", id).
			First(info).Error
	return info, err
}

func (e *Engine) getMovieInfoWithCallback(provider mt.MovieProvider, id string, lazy bool, callback func() (*model.MovieInfo, error)) (info *model.MovieInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.Valid()) {
			err = mt.ErrIncompleteMetadata
		}
	}()
	// Query DB first (by id).
	if lazy {
		if info, err = e.getMovieInfoFromDB(provider, id); err == nil && info.Valid() {
			return // ignore DB query error.
		}
	}
	// delayed info auto-save.
	defer func() {
		if err == nil && info.Valid() {
			e.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(info) // ignore error
		}
	}()
	return callback()
}

func (e *Engine) getMovieInfoByProviderID(provider mt.MovieProvider, id string, lazy bool) (*model.MovieInfo, error) {
	if id = provider.NormalizeID(id); id == "" {
		return nil, mt.ErrInvalidID
	}
	return e.getMovieInfoWithCallback(provider, id, lazy, func() (*model.MovieInfo, error) {
		return provider.GetMovieInfoByID(id)
	})
}

func (e *Engine) GetMovieInfoByProviderID(name, id string, lazy bool) (*model.MovieInfo, error) {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.getMovieInfoByProviderID(provider, id, lazy)
}

func (e *Engine) getMovieInfoByProviderURL(provider mt.MovieProvider, rawURL string, lazy bool) (*model.MovieInfo, error) {
	id, err := provider.ParseIDFromURL(rawURL)
	switch {
	case err != nil:
		return nil, err
	case id == "":
		return nil, mt.ErrInvalidURL
	}
	return e.getMovieInfoWithCallback(provider, id, lazy, func() (*model.MovieInfo, error) {
		return provider.GetMovieInfoByURL(rawURL)
	})
}

func (e *Engine) GetMovieInfoByURL(rawURL string, lazy bool) (*model.MovieInfo, error) {
	provider, err := e.GetMovieProviderByURL(rawURL)
	if err != nil {
		return nil, err
	}
	return e.getMovieInfoByProviderURL(provider, rawURL, lazy)
}
