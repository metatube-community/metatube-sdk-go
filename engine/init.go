package engine

import (
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

// Special environment prefixes for setting provider priorities.
const (
	ActorProviderPriorityEnvPrefix = "MT_ACTOR_PROVIDER_PRIORITY_"
	MovieProviderPriorityEnvPrefix = "MT_MOVIE_PROVIDER_PRIORITY_"
)

func (e *Engine) init() *Engine {
	e.fetcher = fetch.Default(&fetch.Config{Timeout: e.timeout})
	e.initLogger()
	e.initActorProviders(e.timeout)
	e.initMovieProviders(e.timeout)
	e.initAllProviderPriorities()
	return e
}

func (e *Engine) initLogger() {
	logConf := zap.NewProductionConfig()
	logConf.Encoding = "console"
	logger, _ := logConf.Build()
	e.logger = logger.Sugar()
}

func (e *Engine) initAllProviderPriorities() {
	for _, env := range os.Environ() {
		key, value, _ := strings.Cut(strings.ToUpper(env), "=")
		switch {
		case strings.HasPrefix(key, ActorProviderPriorityEnvPrefix):
			name := key[len(ActorProviderPriorityEnvPrefix):]
			prio, _ := strconv.ParseInt(value, 0, 64)
			if prio == 0 {
				delete(e.actorProviders, name)
				continue
			}
			if provider, ok := e.actorProviders[name]; ok {
				provider.SetPriority(prio)
			}
		case strings.HasPrefix(key, MovieProviderPriorityEnvPrefix):
			name := key[len(MovieProviderPriorityEnvPrefix):]
			prio, _ := strconv.ParseInt(value, 0, 64)
			if prio == 0 {
				delete(e.movieProviders, name)
				continue
			}
			if provider, ok := e.movieProviders[name]; ok {
				provider.SetPriority(prio)
			}
		}
	}
}

// initActorProviders initializes actor providers.
func (e *Engine) initActorProviders(timeout time.Duration) {
	{ // init
		e.actorProviders = make(map[string]mt.ActorProvider)
		e.actorHostProviders = make(map[string][]mt.ActorProvider)
	}
	mt.RangeActorFactory(func(name string, factory mt.ActorFactory) {
		provider := factory()
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		// Add actor provider by name.
		e.actorProviders[strings.ToUpper(name)] = provider
		// Add actor provider by host.
		host := provider.URL().Hostname()
		e.actorHostProviders[host] = append(e.actorHostProviders[host], provider)
	})
}

// initMovieProviders initializes movie providers.
func (e *Engine) initMovieProviders(timeout time.Duration) {
	{ // init
		e.movieProviders = make(map[string]mt.MovieProvider)
		e.movieHostProviders = make(map[string][]mt.MovieProvider)
	}
	mt.RangeMovieFactory(func(name string, factory mt.MovieFactory) {
		provider := factory()
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		// Add movie provider by name.
		e.movieProviders[strings.ToUpper(name)] = provider
		// Add movie provider by host.
		host := provider.URL().Hostname()
		e.movieHostProviders[host] = append(e.movieHostProviders[host], provider)
	})
}
