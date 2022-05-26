package engine

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

type Engine struct {
	db             *gorm.DB
	actorProviders map[string]javtube.ActorProvider
	movieProviders map[string]javtube.MovieProvider
}

func New(db *gorm.DB, timeout time.Duration) *Engine {
	engine := &Engine{db: db}
	engine.initActorProviders(timeout)
	engine.initMovieProviders(timeout)
	return engine
}

// initActorProviders initializes actor providers.
func (e *Engine) initActorProviders(timeout time.Duration) {
	e.actorProviders = make(map[string]javtube.ActorProvider)
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		e.actorProviders[strings.ToUpper(name)] = factory()
	})
	return
}

// initMovieProviders initializes movie providers.
func (e *Engine) initMovieProviders(timeout time.Duration) {
	e.movieProviders = make(map[string]javtube.MovieProvider)
	javtube.RangeMovieFactory(func(name string, factory javtube.MovieFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		e.movieProviders[strings.ToUpper(name)] = provider
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

func (e *Engine) GetActorProviderByURL(rawURL string) (javtube.ActorProvider, error) {
	return nil, nil
}

func (e *Engine) GetActorProviderByName(name string) (javtube.ActorProvider, error) {
	provider, ok := e.actorProviders[strings.ToUpper(name)]
	if !ok {
		return nil, fmt.Errorf("actor provider not found: %s", name)
	}
	return provider, nil
}

func (e *Engine) MustGetActorProviderByName(name string) javtube.ActorProvider {
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

func (e *Engine) GetMovieProviderByURL(rawURL string) (javtube.MovieProvider, error) {
	return nil, nil
}

func (e *Engine) GetMovieProviderByName(name string) (javtube.MovieProvider, error) {
	provider, ok := e.movieProviders[strings.ToUpper(name)]
	if !ok {
		return nil, fmt.Errorf("movie provider not found: %s", name)
	}
	return provider, nil
}

func (e *Engine) MustGetMovieProviderByName(name string) javtube.MovieProvider {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		panic(err)
	}
	return provider
}
