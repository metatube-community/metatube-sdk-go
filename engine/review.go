package engine

import (
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *Engine) getMovieReviewsFromDB(provider mt.MovieProvider, id string) (*model.MovieReviewInfo, error) {
	info := &model.MovieReviewInfo{}
	err := e.db. // Exact match here.
			Where("provider = ?", provider.Name()).
			Where("id = ? COLLATE NOCASE", id).
			First(info).Error
	return info, err
}

func (e *Engine) getMovieReviewsWithCallback(provider mt.MovieProvider, id string, lazy bool,
	callback func() ([]*model.MovieReviewDetail, error),
) (info *model.MovieReviewInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.IsValid()) {
			err = mt.ErrIncompleteMetadata
		}
	}()
	// Query DB first (by id).
	if lazy {
		if info, err = e.getMovieReviewsFromDB(provider, id); err == nil && info.IsValid() {
			return // ignore DB query error.
		}
	}
	// delayed info auto-save.
	defer func() {
		if err == nil && info.IsValid() {
			e.db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(info) // ignore error
		}
	}()

	var reviews []*model.MovieReviewDetail
	if reviews, err = callback(); err != nil {
		return
	}

	info = &model.MovieReviewInfo{
		ID:       id,
		Provider: provider.Name(),
		Reviews:  datatypes.NewJSONType(reviews),
	}
	return
}

func (e *Engine) getMovieReviewsByProviderID(provider mt.MovieProvider, id string, lazy bool) (*model.MovieReviewInfo, error) {
	if id = provider.NormalizeMovieID(id); id == "" {
		return nil, mt.ErrInvalidID
	}

	reviewer, ok := provider.(mt.MovieReviewer)
	if !ok {
		return nil, fmt.Errorf("reviews not supported by %s", provider.Name())
	}

	return e.getMovieReviewsWithCallback(provider, id, lazy, func() ([]*model.MovieReviewDetail, error) {
		return reviewer.GetMovieReviewsByID(id)
	})
}

func (e *Engine) GetMovieReviewsByProviderID(pid providerid.ProviderID, lazy bool) (*model.MovieReviewInfo, error) {
	provider, err := e.GetMovieProviderByName(pid.Provider)
	if err != nil {
		return nil, err
	}
	return e.getMovieReviewsByProviderID(provider, pid.ID, lazy)
}

func (e *Engine) getMovieReviewsByProviderURL(provider mt.MovieProvider, rawURL string, lazy bool) (*model.MovieReviewInfo, error) {
	id, err := provider.ParseMovieIDFromURL(rawURL)
	switch {
	case err != nil:
		return nil, err
	case id == "":
		return nil, mt.ErrInvalidURL
	}

	reviewer, ok := provider.(mt.MovieReviewer)
	if !ok {
		return nil, fmt.Errorf("reviews not supported by %s", provider.Name())
	}

	return e.getMovieReviewsWithCallback(provider, id, lazy, func() ([]*model.MovieReviewDetail, error) {
		return reviewer.GetMovieReviewsByURL(rawURL)
	})
}

func (e *Engine) GetMovieReviewsByProviderURL(rawURL string, lazy bool) (*model.MovieReviewInfo, error) {
	provider, err := e.GetMovieProviderByURL(rawURL)
	if err != nil {
		return nil, err
	}
	return e.getMovieReviewsByProviderURL(provider, rawURL, lazy)
}
