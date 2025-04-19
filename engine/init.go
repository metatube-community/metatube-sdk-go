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
	e.actorProviders = make(map[string]mt.ActorProvider)
	e.actorHostProviders = make(map[string][]mt.ActorProvider)
	for name, factory := range mt.RangeActorFactory {
		name = strings.ToUpper(name)

		provider := factory()
		if p, ok := e.actorConfigManager.GetPriority(name); ok {
			e.logger.Printf("Set actor provider with overridden priority: %s=%.2f", provider.Name(), p)
			provider.SetPriority(p)
		}
		if provider.Priority() <= 0 {
			e.logger.Printf("Disable actor provider: %s", provider.Name())
			continue
		}

		// Set request timeout.
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			timeout := e.timeout
			if v, ok := e.actorConfigManager.GetTimeout(name); ok {
				e.logger.Printf("Set actor provider with overridden request timeout: %s=%s", provider.Name(), v)
				timeout = v // override global timeout.
			}
			s.SetRequestTimeout(timeout)
		}

		// Set provider config.
		if s, ok := provider.(mt.ConfigSetter); ok {
			if err := s.SetConfig(e.actorConfigManager.GetConfig(name)); err != nil {
				e.logger.Fatalf("Set actor provider config for %s: %v", provider.Name(), err)
			}
		}

		// Add actor provider by name.
		e.actorProviders[name] = provider
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
		name = strings.ToUpper(name)

		provider := factory()
		if p, ok := e.movieConfigManager.GetPriority(name); ok {
			e.logger.Printf("Set movie provider with overridden priority: %s=%.2f", provider.Name(), p)
			provider.SetPriority(p)
		}
		if provider.Priority() <= 0 {
			e.logger.Printf("Disable movie provider: %s", provider.Name())
			continue
		}

		// Set request timeout.
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			timeout := e.timeout
			if v, ok := e.movieConfigManager.GetTimeout(name); ok {
				e.logger.Printf("Set movie provider with overridden request timeout: %s=%s", provider.Name(), v)
				timeout = v // override global timeout.
			}
			s.SetRequestTimeout(timeout)
		}

		// Set provider config.
		if s, ok := provider.(mt.ConfigSetter); ok {
			if err := s.SetConfig(e.movieConfigManager.GetConfig(name)); err != nil {
				e.logger.Fatalf("Set movie provider config for %s: %v", provider.Name(), err)
			}
		}

		// Add movie provider by name.
		e.movieProviders[name] = provider
		// Add movie provider by host.
		host := provider.URL().Hostname()
		e.movieHostProviders[host] = append(e.movieHostProviders[host], provider)
	}
}
