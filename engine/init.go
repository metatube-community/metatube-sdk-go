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
	e.initActorImageProviders()
	e.initMovieProviders()
	return e
}

func (e *Engine) initLogger() {
	e.logger = log.New(os.Stdout, "[ENGINE]\u0020", log.LstdFlags|log.Llongfile)
}

func (e *Engine) initFetcher() {
	e.fetcher = fetch.Default(&fetch.Config{Timeout: e.timeout})
}

// initActorProviders initializes actor providers.
func (e *Engine) initActorProviders() {
	defer func() {
		// remove references.
		e.actorPriorities = nil
	}()

	e.actorProviders = make(map[string]mt.ActorProvider)
	e.actorHostProviders = make(map[string][]mt.ActorProvider)
	for name, factory := range mt.RangeActorFactory {
		name = strings.ToUpper(name)

		provider := factory()
		if p, ok := e.actorPriorities[name]; ok {
			e.logger.Printf("Set actor provider with overridden priority: %s=%.2f", provider.Name(), p)
			provider.SetPriority(p)
		}
		if provider.Priority() <= 0 {
			e.logger.Printf("Disable actor provider: %s", provider.Name())
			continue
		}

		// Set request timeout.
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(e.timeout)
		}

		// Add actor provider by name.
		e.actorProviders[name] = provider
		// Add actor provider by host.
		host := provider.URL().Hostname()
		e.actorHostProviders[host] = append(e.actorHostProviders[host], provider)
	}
}

func (e *Engine) initActorImageProviders() {
	defer func() {
		// remove references.
		e.actorImagePriorities = nil
	}()

	e.actorImageProviders = make(map[string]mt.ActorImageProvider)
	e.actorImageLanguageProviders = make(map[string][]mt.ActorImageProvider)
	for name, factory := range mt.RangeActorImageFactory {
		name = strings.ToUpper(name)

		provider := factory()
		if p, ok := e.actorImagePriorities[name]; ok {
			e.logger.Printf("Set actor image provider with overridden priority: %s=%.2f", provider.Name(), p)
			provider.SetPriority(p)
		}
		if provider.Priority() <= 0 {
			e.logger.Printf("Disable actor image provider: %s", provider.Name())
			continue
		}

		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(e.timeout)
		}
		// Add actor image provider by name.
		e.actorImageProviders[strings.ToUpper(name)] = provider
		// Add actor image provider by language.
		lang := provider.Language().String()
		e.actorImageLanguageProviders[lang] = append(e.actorImageLanguageProviders[lang], provider)
	}
}

// initMovieProviders initializes movie providers.
func (e *Engine) initMovieProviders() {
	defer func() {
		// remove references.
		e.moviePriorities = nil
	}()

	e.movieProviders = make(map[string]mt.MovieProvider)
	e.movieHostProviders = make(map[string][]mt.MovieProvider)
	for name, factory := range mt.RangeMovieFactory {
		name = strings.ToUpper(name)

		provider := factory()
		if p, ok := e.moviePriorities[name]; ok {
			e.logger.Printf("Set movie provider with overridden priority: %s=%.2f", provider.Name(), p)
			provider.SetPriority(p)
		}
		if provider.Priority() <= 0 {
			e.logger.Printf("Disable movie provider: %s", provider.Name())
			continue
		}

		// Set request timeout.
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(e.timeout)
		}

		// Add movie provider by name.
		e.movieProviders[name] = provider
		// Add movie provider by host.
		host := provider.URL().Hostname()
		e.movieHostProviders[host] = append(e.movieHostProviders[host], provider)
	}
}
