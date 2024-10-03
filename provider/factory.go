package provider

import (
	"fmt"
	"sync"
)

type (
	MovieFactory = func() MovieProvider
	ActorFactory = func() ActorProvider
)

var (
	// Factory RW Mutex
	factoryMu sync.RWMutex
	// Actor/Movie Factories
	movieFactories = make(map[string]MovieFactory)
	actorFactories = make(map[string]ActorFactory)
)

func Register[T Provider](name string, factory func() T) {
	factoryMu.Lock()
	defer factoryMu.Unlock()

	registered := false
	provider := *new(T)

	if _, ok := any(provider).(ActorProvider); ok {
		actorFactories[name] = func() ActorProvider { return any(factory()).(ActorProvider) }
		registered = true
	}

	if _, ok := any(provider).(MovieProvider); ok {
		movieFactories[name] = func() MovieProvider { return any(factory()).(MovieProvider) }
		registered = true
	}

	if !registered {
		panic(fmt.Sprintf("invalid provider factory: func() %T", provider))
	}
}

func RangeMovieFactory(f func(string, MovieFactory)) {
	factoryMu.RLock()
	for name, factory := range movieFactories {
		f(name, factory)
	}
	factoryMu.RUnlock()
}

func RangeActorFactory(f func(string, ActorFactory)) {
	factoryMu.RLock()
	for name, factory := range actorFactories {
		f(name, factory)
	}
	factoryMu.RUnlock()
}
