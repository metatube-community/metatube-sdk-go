package provider

import (
	"reflect"
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

	// Get the return type of the factory function.
	t := reflect.TypeOf(factory).Out(0)

	// Track if the factory has been registered.
	registered := false

	// Check if the return type implements ActorProvider.
	if t.Implements(reflect.TypeOf((*ActorProvider)(nil)).Elem()) {
		actorFactories[name] = func() ActorProvider { return any(factory()).(ActorProvider) }
		registered = true
	}

	// Check if the return type implements MovieProvider.
	if t.Implements(reflect.TypeOf((*MovieProvider)(nil)).Elem()) {
		movieFactories[name] = func() MovieProvider { return any(factory()).(MovieProvider) }
		registered = true
	}

	// Panic if the factory does not implement either interface.
	if !registered {
		panic("invalid provider factory: func() " + t.String())
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
