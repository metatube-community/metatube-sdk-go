package dbengine

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/model"
)

var _ DBEngine = (*engine)(nil)

type DBEngine interface {
	actorEngine
	movieEngine
	AutoMigrate() error
	Driver() string
	Version() (string, error)
}

type engine struct {
	db *gorm.DB
}

func New(db *gorm.DB) DBEngine {
	return &engine{db: db}
}

func (e *engine) DB() *gorm.DB {
	return e.db.Session(&gorm.Session{})
}

func (e *engine) Driver() string {
	return e.db.Config.Dialector.Name()
}

func (e *engine) AutoMigrate() error {
	if e.Driver() == database.Postgres {
		sqlStmts := []string{
			// Create case-insensitive collation.
			`CREATE COLLATION IF NOT EXISTS nocase (
			   provider = icu,
			   locale = 'und-u-ks-level2',
			   deterministic = FALSE
			 )`,
			// Create pg_trgm extension.
			`CREATE EXTENSION IF NOT EXISTS pg_trgm`,
		}
		for _, sql := range sqlStmts {
			if err := e.db.Exec(sql).Error; err != nil {
				return err
			}
		}
	}

	// Table auto migration.
	if err := e.db.AutoMigrate(
		&model.MovieInfo{},
		&model.ActorInfo{},
		&model.MovieReviewInfo{},
	); err != nil {
		return err
	}

	if e.Driver() == database.Postgres {
		buildNocaseIndexSQL := func(table, column string) string {
			const tmpl = `CREATE INDEX IF NOT EXISTS idx_%s_%s_nocase ON %s (%s COLLATE nocase)`
			return fmt.Sprintf(tmpl, table, column, table, column)
		}
		buildTrgmIndexSQL := func(table, column string) string {
			const tmpl = `CREATE INDEX IF NOT EXISTS idx_%s_%s_trgm ON %s USING gin (%s gin_trgm_ops)`
			return fmt.Sprintf(tmpl, table, column, table, column)
		}
		sqlStmts := []string{
			// Create indexes for nocase collation.
			buildNocaseIndexSQL(model.ActorMetadataTableName, "provider"),
			buildNocaseIndexSQL(model.ActorMetadataTableName, "id"),
			buildNocaseIndexSQL(model.ActorMetadataTableName, "name"),
			buildNocaseIndexSQL(model.MovieMetadataTableName, "provider"),
			buildNocaseIndexSQL(model.MovieMetadataTableName, "id"),
			buildNocaseIndexSQL(model.MovieMetadataTableName, "number"),
			// Create indexes for full-text search.
			buildTrgmIndexSQL(model.ActorMetadataTableName, "name"),
			buildTrgmIndexSQL(model.MovieMetadataTableName, "number"),
			buildTrgmIndexSQL(model.MovieMetadataTableName, "title"),
		}
		for _, sql := range sqlStmts {
			if err := e.db.Exec(sql).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *engine) Version() (version string, err error) {
	switch e.Driver() {
	case database.Postgres:
		err = e.DB().Raw("SELECT version();").Scan(&version).Error
	case database.Sqlite:
		err = e.DB().Raw("SELECT sqlite_version();").Scan(&version).Error
	default:
		err = fmt.Errorf("unsupported DB type: %s", e.Driver())
	}
	return
}
