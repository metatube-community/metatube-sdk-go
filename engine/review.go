package engine

import (
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *Engine) getMovieReviewsFromDB(provider mt.MovieProvider, id string) (*model.MovieReviews, error) {
	info := &model.MovieReviews{}
	err := e.db. // Exact match here.
			Where("provider = ?", provider.Name()).
			Where("id = ? COLLATE NOCASE", id).
			First(info).Error
	return info, err
}

func (e *Engine) getMovieReviewsWithCallback(provider mt.MovieProvider, id string, lazy bool, callback func() ([]*model.MovieReviewInfo, error)) (info *model.MovieReviews, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.Valid()) {
			err = mt.ErrIncompleteMetadata
		}
	}()
	// Query DB first (by id).
	if lazy {
		if info, err = e.getMovieReviewsFromDB(provider, id); err == nil && info.Valid() {
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

	var reviews []*model.MovieReviewInfo
	if reviews, err = callback(); err != nil {
		return
	}

	info = &model.MovieReviews{
		ID:       id,
		Provider: provider.Name(),
		Reviews:  datatypes.NewJSONType(reviews),
	}
	return
}

func (e *Engine) getMovieReviewsByProviderID(provider mt.MovieProvider, id string, lazy bool) (*model.MovieReviews, error) {
	if id = provider.NormalizeMovieID(id); id == "" {
		return nil, mt.ErrInvalidID
	}

	reviewer, ok := provider.(mt.MovieReviewer)
	if !ok {
		return nil, fmt.Errorf("reviews not supported by %s", provider.Name())
	}

	return e.getMovieReviewsWithCallback(provider, id, lazy, func() ([]*model.MovieReviewInfo, error) {
		return reviewer.GetMovieReviewsByID(id)
	})
}

func (e *Engine) GetMovieReviewsByProviderID(name, id string, lazy bool) (*model.MovieReviews, error) {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.getMovieReviewsByProviderID(provider, id, lazy)
}

func (e *Engine) getMovieReviewsByProviderURL(provider mt.MovieProvider, rawURL string, lazy bool) (*model.MovieReviews, error) {
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

	return e.getMovieReviewsWithCallback(provider, id, lazy, func() ([]*model.MovieReviewInfo, error) {
		return reviewer.GetMovieReviewsByURL(rawURL)
	})
}

func (e *Engine) GetMovieReviewsByProviderURL(name, rawURL string, lazy bool) (*model.MovieReviews, error) {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		return nil, err
	}
	return e.getMovieReviewsByProviderURL(provider, rawURL, lazy)
}
