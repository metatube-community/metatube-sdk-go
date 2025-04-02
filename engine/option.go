package engine

import (
	"time"
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

func WithActorProviderPriorities(priorities map[string]float64) Option {
	return func(e *Engine) {
		e.actorPriorities = priorities
	}
}

func WithMovieProviderPriorities(priorities map[string]float64) Option {
	return func(e *Engine) {
		e.moviePriorities = priorities
	}
}
