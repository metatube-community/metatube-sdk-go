package dbengine

import (
	"errors"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
)

type movieEngine interface {
	GetMovieInfo(providerid.ProviderID) (*model.MovieInfo, error)
	SaveMovieInfo(*model.MovieInfo) error
	SearchMovie(string, SearchOptions) ([]*model.MovieSearchResult, error)

	GetMovieReviewInfo(providerid.ProviderID) (*model.MovieReviewInfo, error)
	SaveMovieReviewInfo(*model.MovieReviewInfo) error
}

var _ movieEngine = (*engine)(nil)

func (e *engine) GetMovieInfo(pid providerid.ProviderID) (*model.MovieInfo, error) {
	info := &model.MovieInfo{}
	err := e.DB().
		Where( // Exact match here.
			`provider COLLATE NOCASE = ? AND id COLLATE NOCASE = ?`,
			pid.Provider, pid.ID,
		).First(info).Error
	return info, err
}

func (e *engine) SaveMovieInfo(info *model.MovieInfo) error {
	if !info.IsValid() {
		return fmt.Errorf("invalid %T", info)
	}
	return e.DB().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(info).Error
}

func (e *engine) SearchMovie(keyword string, opts SearchOptions) ([]*model.MovieSearchResult, error) {
	opts.applyDefaults()

	// DB session.
	tx := e.DB()

	// provider filter.
	if opts.Provider != "" {
		tx = tx.Where(`provider COLLATE NOCASE = ?`, opts.Provider)
	}

	// keyword filter.
	tx = tx.Where(
		// Note: keyword might be an ID or just a regular number, so we should
		// query both of them for a better match. Also, it's case-insensitive.
		`(number COLLATE NOCASE = ? OR id COLLATE NOCASE = ?)`,
		keyword, keyword,
	)

	// pagination.
	if opts.Limit > 0 {
		tx = tx.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		tx = tx.Offset(opts.Offset)
	}

	var infos []*model.MovieInfo
	if err := tx.Find(&infos).Error; err != nil {
		return nil, err
	}

	results := make([]*model.MovieSearchResult, 0, len(infos))
	for _, info := range infos {
		if !info.IsValid() {
			continue // normally it is valid, but just in case.
		}
		results = append(results, info.ToSearchResult())
	}
	return results, nil
}

func (e *engine) GetMovieReviewInfo(pid providerid.ProviderID) (*model.MovieReviewInfo, error) {
	info := &model.MovieReviewInfo{}
	err := e.DB().
		Where( // Exact match here.
			`provider COLLATE NOCASE = ? AND id COLLATE NOCASE = ?`,
			pid.Provider, pid.ID,
		).First(info).Error
	return info, err
}

func (e *engine) SaveMovieReviewInfo(info *model.MovieReviewInfo) error {
	if !info.IsValid() {
		return fmt.Errorf("invalid %T", info)
	}
	if len(info.Reviews.Data()) == 0 {
		return errors.New("reviews cannot be empty")
	}
	return e.DB().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(info).Error
}
