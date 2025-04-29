package dbengine

import (
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *engine) SearchMovie(provider mt.Provider, keyword string) ([]*model.MovieSearchResult, error) {
	var infos []*model.MovieInfo
	tx := e.DB()
	if provider != nil {
		tx = tx.Where(`provider COLLATE NOCASE = ?`, provider.Name())
	}
	tx = tx.Where(
		// Note: keyword might be an ID or just a regular number, so we should
		// query both of them for a better match. Also, it's case-insensitive.
		`(number COLLATE NOCASE = ? OR id COLLATE NOCASE = ?)`,
		keyword, keyword,
	)
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

func (e *engine) GetMovieInfo(provider mt.Provider, id string) (*model.MovieInfo, error) {
	info := &model.MovieInfo{}
	err := e.DB().
		Where( // Exact match here.
			`provider COLLATE NOCASE = ? AND id COLLATE NOCASE = ?`,
			provider.Name(), id,
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

func (e *engine) GetMovieReviewInfo(provider mt.Provider, id string) (*model.MovieReviewInfo, error) {
	info := &model.MovieReviewInfo{}
	err := e.DB().
		Where( // Exact match here.
			`provider COLLATE NOCASE = ? AND id COLLATE NOCASE = ?`,
			provider.Name(), id,
		).First(info).Error
	return info, err
}

func (e *engine) SaveMovieReviewInfo(info *model.MovieReviewInfo) error {
	if !info.IsValid() {
		return fmt.Errorf("invalid %T", info)
	}
	return e.DB().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(info).Error
}
