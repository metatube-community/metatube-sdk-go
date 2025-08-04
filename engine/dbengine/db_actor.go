package dbengine

import (
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
)

type actorEngine interface {
	GetActorInfo(providerid.ProviderID) (*model.ActorInfo, error)
	SaveActorInfo(*model.ActorInfo) error
	SearchActor(string, ActorSearchOptions) ([]*model.ActorSearchResult, error)
}

var _ actorEngine = (*engine)(nil)

func (e *engine) GetActorInfo(pid providerid.ProviderID) (*model.ActorInfo, error) {
	info := &model.ActorInfo{}
	err := e.DB().
		Where( // Exact match here.
			`provider COLLATE NOCASE = ? AND id COLLATE NOCASE = ?`,
			pid.Provider, pid.ID,
		).First(info).Error
	return info, err
}

func (e *engine) SaveActorInfo(info *model.ActorInfo) error {
	if !info.IsValid() {
		return fmt.Errorf("invalid %T", info)
	}
	return e.DB().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(info).Error
}

func (e *engine) SearchActor(keyword string, opts ActorSearchOptions) ([]*model.ActorSearchResult, error) {
	opts.applyDefaults()

	// DB session.
	tx := e.DB()

	// provider filter.
	if opts.Provider != "" {
		tx = tx.Where(`provider COLLATE NOCASE = ?`, opts.Provider)
	}

	// keyword filter.
	if e.Driver() == database.Postgres {
		tx = tx.Where(
			`(name COLLATE NOCASE = ? OR similarity(name, ?) > ?)`,
			keyword, keyword, opts.Threshold,
		)
	} else { // Sqlite
		pattern := "%" + keyword + "%"
		tx = tx.Where(
			`(name COLLATE NOCASE = ? OR name LIKE ? COLLATE NOCASE)`,
			keyword, pattern,
		)
	}

	// pagination.
	if opts.Limit > 0 {
		tx = tx.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		tx = tx.Offset(opts.Offset)
	}

	var infos []*model.ActorInfo
	if err := tx.Find(&infos).Error; err != nil {
		return nil, err
	}

	results := make([]*model.ActorSearchResult, 0, len(infos))
	for _, info := range infos {
		if !info.IsValid() {
			continue // ignore invalid info.
		}
		results = append(results, info.ToSearchResult())
	}
	return results, nil
}
