package provider

import (
	"sync"
)

type (
	Factory      = func() Provider
	ActorFactory = func() ActorProvider
)

var (
	// Mutexes
	factoryMu      sync.Mutex
	actorFactoryMu sync.Mutex
	// Factories
	factories      = make(map[string]Factory)
	actorFactories = make(map[string]ActorFactory)
)

func RegisterFactory(name string, f Factory) {
	factoryMu.Lock()
	factories[name] = f
	factoryMu.Unlock()
}

func RangeFactory(f func(string, Factory)) {
	factoryMu.Lock()
	for name, factory := range factories {
		f(name, factory)
	}
	factoryMu.Unlock()
}

func RegisterActorFactory(name string, f ActorFactory) {
	actorFactoryMu.Lock()
	actorFactories[name] = f
	actorFactoryMu.Unlock()
}

func RangeActorFactory(f func(string, ActorFactory)) {
	actorFactoryMu.Lock()
	for name, factory := range actorFactories {
		f(name, factory)
	}
	actorFactoryMu.Unlock()
}
