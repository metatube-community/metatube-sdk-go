package provider

import (
	"sync"
)

type (
	Factory      = func() Provider
	ActorFactory = func() ActorProvider
)

var (
	// RW Mutexes
	factoryMu      sync.RWMutex
	actorFactoryMu sync.RWMutex
	// Factories
	factories      = make(map[string]Factory)
	actorFactories = make(map[string]ActorFactory)
)

func RegisterFactory[T Provider](name string, factory func() T) {
	factoryMu.Lock()
	factories[name] = func() Provider { return factory() }
	factoryMu.Unlock()
}

func RangeFactory(f func(string, Factory)) {
	factoryMu.RLock()
	for name, factory := range factories {
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
