package engine

import (
	"sync"

	"gorm.io/gorm/clause"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/priority"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

func (e *Engine) searchMovie(keyword string, provider javtube.MovieProvider, lazy bool) ([]*model.MovieSearchResult, error) {
	// Regular keyword searching.
	if searcher, ok := provider.(javtube.MovieSearcher); ok {
		if keyword = searcher.TidyKeyword(keyword); keyword == "" {
			return nil, javtube.ErrInvalidKeyword
		}
		// Query DB first (by number).
		if info := new(model.MovieInfo); lazy {
			if result := e.db.
				Where("provider = ?", provider.Name()).
				Where(e.db.
					// Exact match.
					Where("number = ?", keyword).
					Or("id = ?", keyword)).
				First(info); result.Error == nil && info.Valid() /* must be valid */ {
				return []*model.MovieSearchResult{info.ToSearchResult()}, nil
			} // ignore DB query error.
		}
		return searcher.SearchMovie(keyword)
	}
	// Fallback to movie info querying.
	info, err := e.getMovieInfoByID(keyword, provider, true)
	if err != nil {
		return nil, err
	}
	return []*model.MovieSearchResult{info.ToSearchResult()}, nil
}

func (e *Engine) SearchMovie(keyword, name string, lazy bool) ([]*model.MovieSearchResult, error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.searchMovie(keyword, provider, lazy)
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
		go func(provider javtube.MovieProvider) {
			defer wg.Done()
			results, err := e.searchMovie(keyword, provider, false)
			respCh <- response{
				Results: results,
				Error:   err,
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
func (e *Engine) SearchMovieAll(keyword string, lazy bool) (results []*model.MovieSearchResult, err error) {
	if keyword = number.Trim(keyword); keyword == "" {
		return nil, javtube.ErrInvalidKeyword
	}

	defer func() {
		if err != nil {
			return
		}
		if len(results) == 0 {
			err = javtube.ErrNotFound
			return
		}
		// post-processing
		ps := new(priority.Slice[float64, *model.MovieSearchResult])
		for _, result := range results {
			if !result.Valid() /* validation check */ {
				continue
			}
			ps.Append(number.Similarity(keyword, result.Number)*
				float64(e.MustGetMovieProviderByName(result.Provider).Priority()), result)
		}
		// sort according to priority.
		results = ps.Sort().Underlying()
	}()

	if lazy {
		multiInfo := make([]*model.MovieInfo, 0)
		if result := e.db.
			// Note: keyword might be an ID or just a regular number, so we should
			// query both of them for best match. Also, case should not mater.
			Where("UPPER(number) = UPPER(?)", keyword).
			Or("UPPER(id) = UPPER(?)", keyword).
			Find(&multiInfo); result.Error == nil && result.RowsAffected > 0 {
			for _, info := range multiInfo {
				results = append(results, info.ToSearchResult())
			}
			return
		}
	}

	results, err = e.searchMovieAll(keyword)
	return
}

func (e *Engine) getMovieInfoFromDB(id string, provider javtube.MovieProvider) (*model.MovieInfo, error) {
	info := &model.MovieInfo{}
	err := e.db. // Exact match here.
			Where("id = ?", id).
			Where("provider = ?", provider.Name()).
			First(info).Error
	return info, err
}

func (e *Engine) getMovieInfoByID(id string, provider javtube.MovieProvider, lazy bool) (info *model.MovieInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.Valid()) {
			err = javtube.ErrInvalidMetadata
		}
	}()
	if id = provider.NormalizeID(id); id == "" {
		return nil, javtube.ErrInvalidID
	}
	// Query DB first (by id).
	if lazy {
		if info, err = e.getMovieInfoFromDB(id, provider); err == nil && info.Valid() {
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
	return provider.GetMovieInfoByID(id)
}

func (e *Engine) GetMovieInfoByID(id, name string, lazy bool) (*model.MovieInfo, error) {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.getMovieInfoByID(id, provider, lazy)
}
