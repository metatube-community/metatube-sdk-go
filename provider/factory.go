package provider

import (
	"sync"
)

type (
	MovieFactory = func() MovieProvider
	ActorFactory = func() ActorProvider
)

var (
	// RW Mutexes
	factoryMu      sync.RWMutex
	actorFactoryMu sync.RWMutex
	// Factories
	movieFactories = make(map[string]MovieFactory)
	actorFactories = make(map[string]ActorFactory)
)

func RegisterMovieFactory[T MovieProvider](name string, factory func() T) {
	factoryMu.Lock()
	movieFactories[name] = func() MovieProvider { return factory() }
	factoryMu.Unlock()
}

func RangeMovieFactory(f func(string, MovieFactory)) {
	factoryMu.RLock()
	for name, factory := range movieFactories {
		f(name, factory)
	}
	factoryMu.RUnlock()
}

func RegisterActorFactory[T ActorProvider](name string, factory func() T) {
	actorFactoryMu.Lock()
	actorFactories[name] = func() ActorProvider { return factory() }
	actorFactoryMu.Unlock()
}

func RangeActorFactory(f func(string, ActorFactory)) {
	actorFactoryMu.RLock()
	for name, factory := range actorFactories {
		f(name, factory)
	}
	actorFactoryMu.RUnlock()
}
