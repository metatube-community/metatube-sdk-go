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

func WithActorProviderConfigs(c mt.ConfigGetter) Option {
	return func(e *Engine) {
		e.actorConfigManager = c
	}
}

func WithMovieProviderConfigs(c mt.ConfigGetter) Option {
	return func(e *Engine) {
		e.movieConfigManager = c
	}
}
