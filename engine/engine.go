package engine

import (
	"strings"
	"time"

	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Options struct {
	// DSN the Data Source Name.
	DSN string

	// Timeout for each request.
	Timeout time.Duration
}

type Engine struct {
	db             *gorm.DB
	movieProviders map[string]javtube.MovieProvider
	actorProviders map[string]javtube.ActorProvider
}

func New(opts *Options) (engine *Engine, err error) {
	var db *gorm.DB
	if db, err = openDB(opts.DSN); err != nil {
		return
	}

	engine = &Engine{
		db:             db,
		movieProviders: make(map[string]javtube.MovieProvider),
		actorProviders: make(map[string]javtube.ActorProvider),
	}
	// Initialize movie providers.
	javtube.RangeMovieFactory(func(name string, factory javtube.MovieFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(opts.Timeout)
		}
		engine.movieProviders[name] = provider
	})
	// Initialize actor providers.
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(opts.Timeout)
		}
		engine.actorProviders[name] = factory()
	})
	return
}

func (e *Engine) AutoMigrate() error {
	return e.db.AutoMigrate(
		&model.MovieInfo{},
		&model.ActorInfo{})
}

func openDB(dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if dsn == "" {
		// We use memory sqlite DB by default.
		dsn = "file::memory:?cache=shared"
	}

	// We try to parse it as postgresql, otherwise
	// fallback to sqlite.
	if strings.HasPrefix(dsn, "postgresql://") ||
		len(strings.Fields(dsn)) > 4 {
		dialector = postgres.Open(dsn)
	} else {
		dialector = sqlite.Open(dsn)
	}

	return gorm.Open(dialector, &gorm.Config{})
}
