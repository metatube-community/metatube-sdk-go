package engine

import (
	"log"
	"os"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func (e *Engine) init() *Engine {
	e.initLogger()
	e.initFetcher()
	e.initActorProviders()
	e.initMovieProviders()
	e.initAllProviderPriorities()
	return e
}

func (e *Engine) initLogger() {
	e.logger = log.New(os.Stdout, "[ENGINE]\u0020", log.LstdFlags|log.Llongfile)
}

func (e *Engine) initFetcher() {
	e.fetcher = fetch.Default(&fetch.Config{Timeout: e.timeout})
}

func (e *Engine) initAllProviderPriorities() {
	defer func() {
		// remove references.
		e.actorPriorities = nil
		e.moviePriorities = nil
	}()
	for name, prio := range e.actorPriorities {
		if prio == 0 {
			delete(e.actorProviders, name)
			continue
		}
		if provider, ok := e.actorProviders[name]; ok {
			provider.SetPriority(prio)
		}
	}
	for name, prio := range e.moviePriorities {
		if prio == 0 {
			delete(e.movieProviders, name)
			continue
		}
		if provider, ok := e.movieProviders[name]; ok {
			provider.SetPriority(prio)
		}
	}
}

// initActorProviders initializes actor providers.
func (e *Engine) initActorProviders() {
	e.actorProviders = make(map[string]mt.ActorProvider)
	e.actorHostProviders = make(map[string][]mt.ActorProvider)
	for name, factory := range mt.RangeActorFactory {
		provider := factory()
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(e.timeout)
		}
		// Add actor provider by name.
		e.actorProviders[strings.ToUpper(name)] = provider
		// Add actor provider by host.
		host := provider.URL().Hostname()
		e.actorHostProviders[host] = append(e.actorHostProviders[host], provider)
	}
}

// initMovieProviders initializes movie providers.
func (e *Engine) initMovieProviders() {
	e.movieProviders = make(map[string]mt.MovieProvider)
	e.movieHostProviders = make(map[string][]mt.MovieProvider)
	for name, factory := range mt.RangeMovieFactory {
		provider := factory()
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(e.timeout)
		}
		// Add movie provider by name.
		e.movieProviders[strings.ToUpper(name)] = provider
		// Add movie provider by host.
		host := provider.URL().Hostname()
		e.movieHostProviders[host] = append(e.movieHostProviders[host], provider)
	}
}
