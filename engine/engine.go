package engine

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/database"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

const (
	DefaultEngineName     = "metatube"
	DefaultRequestTimeout = time.Minute
)

type Engine struct {
	db      *gorm.DB
	name    string
	timeout time.Duration
	fetcher *fetch.Fetcher
	// Engine Logger
	logger *log.Logger
	// Name:Provider Map
	actorProviders map[string]mt.ActorProvider
	movieProviders map[string]mt.MovieProvider
	// Host:Providers Map
	actorHostProviders map[string][]mt.ActorProvider
	movieHostProviders map[string][]mt.MovieProvider
}

func New(db *gorm.DB, opts ...Option) *Engine {
	engine := &Engine{
		db:      db,
		name:    DefaultEngineName,
		timeout: DefaultRequestTimeout,
	}
	// apply options
	for _, opt := range opts {
		opt(engine)
	}
	return engine.init()
}

func Default() *Engine {
	db, _ := database.Open(&database.Config{
		DSN:                  "",
		DisableAutomaticPing: true,
	})
	engine := New(db)
	defer engine.DBAutoMigrate(true)
	return engine
}

func (e *Engine) IsActorProvider(name string) (ok bool) {
	_, ok = e.actorProviders[strings.ToUpper(name)]
	return
}

func (e *Engine) GetActorProviders() map[string]mt.ActorProvider {
	return e.actorProviders
}

func (e *Engine) GetActorProviderByURL(rawURL string) (mt.ActorProvider, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	for _, p := range e.actorHostProviders[u.Hostname()] {
		if strings.HasPrefix(u.Path, p.URL().Path) {
			return p, nil
		}
	}
	return nil, mt.ErrProviderNotFound
}

func (e *Engine) GetActorProviderByName(name string) (mt.ActorProvider, error) {
	provider, ok := e.actorProviders[strings.ToUpper(name)]
	if !ok {
		return nil, mt.ErrProviderNotFound
	}
	return provider, nil
}

func (e *Engine) MustGetActorProviderByName(name string) mt.ActorProvider {
	provider, err := e.GetActorProviderByName(name)
	if err != nil {
		panic(err)
	}
	return provider
}

func (e *Engine) IsMovieProvider(name string) (ok bool) {
	_, ok = e.movieProviders[strings.ToUpper(name)]
	return
}

func (e *Engine) GetMovieProviders() map[string]mt.MovieProvider {
	return e.movieProviders
}

func (e *Engine) GetMovieProviderByURL(rawURL string) (mt.MovieProvider, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	for _, p := range e.movieHostProviders[u.Hostname()] {
		if strings.HasPrefix(u.Path, p.URL().Path) {
			return p, nil
		}
	}
	return nil, mt.ErrProviderNotFound
}

func (e *Engine) GetMovieProviderByName(name string) (mt.MovieProvider, error) {
	provider, ok := e.movieProviders[strings.ToUpper(name)]
	if !ok {
		return nil, mt.ErrProviderNotFound
	}
	return provider, nil
}

func (e *Engine) MustGetMovieProviderByName(name string) mt.MovieProvider {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		panic(err)
	}
	return provider
}

// Fetch fetches content from url. If provider is nil, the
// default fetcher will be used.
func (e *Engine) Fetch(url string, provider mt.Provider) (*http.Response, error) {
	// Provider which implements Fetcher interface should be
	// used to fetch all its corresponding resources.
	if fetcher, ok := provider.(mt.Fetcher); ok {
		return fetcher.Fetch(url)
	}
	return e.fetcher.Fetch(url)
}

// String returns the name of the Engine instance.
func (e *Engine) String() string { return e.name }
