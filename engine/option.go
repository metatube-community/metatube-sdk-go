package engine

import (
	"time"

	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

type Option func(*Engine)

func WithEngineName(name string) Option {
	return func(e *Engine) {
		e.name = name
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(e *Engine) {
		e.timeout = timeout
	}
}

func WithActorProviderConfig(name string, c mt.Config) Option {
	return func(e *Engine) {
		e.actorProviderConfigs.Set(name, c)
	}
}

func WithMovieProviderConfig(name string, c mt.Config) Option {
	return func(e *Engine) {
		e.movieProviderConfigs.Set(name, c)
	}
}
