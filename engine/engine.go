package engine

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

type Options struct {
	// DSN the Data Source Name.
	DSN string

	// DisableAutomaticPing as it is.
	DisableAutomaticPing bool

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

	if !opts.DisableAutomaticPing {
		if pinger, ok := db.ConnPool.(interface{ Ping() error }); ok {
			go pinger.Ping() // Async ping.
		}
	}

	engine = &Engine{
		db:             db,
		actorProviders: initActorProviders(opts.Timeout),
		movieProviders: initMovieProviders(opts.Timeout),
	}
	return
}

// initActorProviders initializes actor providers.
func initActorProviders(timeout time.Duration) (providers map[string]javtube.ActorProvider) {
	providers = make(map[string]javtube.ActorProvider)
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		providers[strings.ToUpper(name)] = factory()
	})
	return
}

// initMovieProviders initializes movie providers.
func initMovieProviders(timeout time.Duration) (providers map[string]javtube.MovieProvider) {
	providers = make(map[string]javtube.MovieProvider)
	javtube.RangeMovieFactory(func(name string, factory javtube.MovieFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		providers[strings.ToUpper(name)] = provider
	})
	return
}

func (e *Engine) AutoMigrate(v bool) error {
	if !v {
		return nil
	}
	return e.db.AutoMigrate(
		&model.MovieInfo{},
		&model.ActorInfo{})
}

func (e *Engine) Fetch(url string, provider javtube.Provider) (*http.Response, error) {
	// Provider which implements Fetcher interface should be
	// used to fetch all its corresponding resources.
	if fetcher, ok := provider.(javtube.Fetcher); ok {
		return fetcher.Fetch(url)
	}
	return fetch.Fetch(url)
}

func (e *Engine) IsActorProvider(name string) (ok bool) {
	_, ok = e.actorProviders[strings.ToUpper(name)]
	return
}

func (e *Engine) GetActorProvider(name string) (javtube.ActorProvider, error) {
	provider, ok := e.actorProviders[strings.ToUpper(name)]
	if !ok {
		return nil, fmt.Errorf("actor provider not found: %s", name)
	}
	return provider, nil
}

func (e *Engine) MustGetActorProvider(name string) javtube.ActorProvider {
	provider, err := e.GetActorProvider(name)
	if err != nil {
		panic(err)
	}
	return provider
}

func (e *Engine) IsMovieProvider(name string) (ok bool) {
	_, ok = e.movieProviders[strings.ToUpper(name)]
	return
}

func (e *Engine) GetMovieProvider(name string) (javtube.MovieProvider, error) {
	provider, ok := e.movieProviders[strings.ToUpper(name)]
	if !ok {
		return nil, fmt.Errorf("movie provider not found: %s", name)
	}
	return provider, nil
}

func (e *Engine) MustGetMovieProvider(name string) javtube.MovieProvider {
	provider, err := e.GetMovieProvider(name)
	if err != nil {
		panic(err)
	}
	return provider
}

func openDB(dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	if dsn == "" {
		// We use memory sqlite DB by default.
		dsn = "file::memory:?cache=shared"
	}

	// We try to parse it as postgresql, otherwise
	// fallback to sqlite.
	if strings.HasPrefix(dsn, "postgres://") ||
		strings.HasPrefix(dsn, "postgresql://") ||
		len(strings.Fields(dsn)) > 4 {
		dialector = postgres.New(postgres.Config{
			DSN: dsn,
			// disables implicit prepared statement usage.
			PreferSimpleProtocol: true,
		})
	} else {
		dialector = sqlite.Open(dsn)
	}

	return gorm.Open(dialector, &gorm.Config{
		DisableAutomaticPing: true,
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             100 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		}),
	})
}
