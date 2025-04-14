package provider

import (
	"fmt"
	"sync"
)

type (
	MovieFactory      = func() MovieProvider
	ActorFactory      = func() ActorProvider
	ActorImageFactory = func() ActorImageProvider
)

var (
	// Factory RW Mutex
	factoryMu sync.RWMutex
	// Actor/Movie Factories
	movieFactories      = make(map[string]MovieFactory)
	actorFactories      = make(map[string]ActorFactory)
	actorImageFactories = make(map[string]ActorImageFactory)
)

func Register[T Provider](name string, factory func() T) {
	factoryMu.Lock()
	defer factoryMu.Unlock()

	provider := *new(T)
	registered := false

	if _, ok := any(provider).(ActorImageProvider); ok {
		actorImageFactories[name] = func() ActorImageProvider { return any(factory()).(ActorImageProvider) }
		registered = true
	}

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

func RangeMovieFactory(f func(string, MovieFactory) bool) {
	factoryMu.RLock()
	for name, factory := range movieFactories {
		if !f(name, factory) {
			return
		}
	}
	factoryMu.RUnlock()
}

func RangeActorFactory(f func(string, ActorFactory) bool) {
	factoryMu.RLock()
	for name, factory := range actorFactories {
		if !f(name, factory) {
			return
		}
	}
	factoryMu.RUnlock()
}

func RangeActorImageFactory(f func(string, ActorImageFactory) bool) {
	factoryMu.RLock()
	for name, factory := range actorImageFactories {
		if !f(name, factory) {
			return
		}
	}
	factoryMu.RUnlock()
}
