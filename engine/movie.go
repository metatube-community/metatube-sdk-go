package engine

import (
	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *Engine) getMovieInfoWithCallback(provider mt.MovieProvider, id string, lazy bool, callback func() (*model.MovieInfo, error)) (info *model.MovieInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.IsValid()) {
			err = mt.ErrIncompleteMetadata
		}
	}()
	// Query DB first (by id).
	if lazy {
		if info, err = e.db.GetMovieInfo(
			providerid.ProviderID{
				Provider: provider.Name(),
				ID:       id,
			}); err == nil && info.IsValid() {
			return // ignore DB query error.
		}
	}
	// delayed info auto-save.
	defer func() {
		if err == nil {
			_ = e.db.SaveMovieInfo(info) // ignore error
		}
	}()
	return callback()
}

func (e *Engine) getMovieInfoByProviderID(provider mt.MovieProvider, id string, lazy bool) (*model.MovieInfo, error) {
	if id = provider.NormalizeMovieID(id); id == "" {
		return nil, mt.ErrInvalidID
	}
	return e.getMovieInfoWithCallback(provider, id, lazy, func() (*model.MovieInfo, error) {
		return provider.GetMovieInfoByID(id)
	})
}

func (e *Engine) GetMovieInfoByProviderID(pid providerid.ProviderID, lazy bool) (*model.MovieInfo, error) {
	provider, err := e.GetMovieProviderByName(pid.Provider)
	if err != nil {
		return nil, err
	}
	return e.getMovieInfoByProviderID(provider, pid.ID, lazy)
}

func (e *Engine) getMovieInfoByProviderURL(provider mt.MovieProvider, rawURL string, lazy bool) (*model.MovieInfo, error) {
	id, err := provider.ParseMovieIDFromURL(rawURL)
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
