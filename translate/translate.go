package translate

import (
	"errors"
	"sync"

	"go.uber.org/atomic"
)

var (
	ErrTranslator = errors.New("translate: unknown translator")
	ErrConfigType = errors.New("translate: invalid config type")
)

type (
	decoderFunc          func(config any) error
	translate[T any]     func(text, from, to string, config T) (string, error)
	configFactory[T any] func() T
	buildConfig[T any]   func(decode decoderFunc) (T, error)
)

type translator struct {
	name        string
	translate   translate[any]
	buildConfig buildConfig[any]
}

// Translators is the list of registered translators.
var (
	translatorsMu     sync.Mutex
	atomicTranslators atomic.Value
)

func Register[T any](name string, tf translate[T], cf configFactory[T]) {
	translatorsMu.Lock()
	translators, _ := atomicTranslators.Load().([]translator)
	atomicTranslators.Store(append(translators, translator{
		name: name,
		translate: func(text, from, to string, config any) (string, error) {
			c, ok := config.(T)
			if !ok {
				return "", ErrConfigType
			}
			return tf(text, from, to, c)
		},
		buildConfig: func(decode decoderFunc) (any, error) {
			c := cf()
			// must pass pointer of the config.
			err := decode(&c)
			return c, err
		},
	}))
	translatorsMu.Unlock()
}

func match(name string) translator {
	translators := atomicTranslators.Load().([]translator)
	for _, t := range translators {
		if t.name == name {
			return t
		}
	}
	return translator{}
}

func BuildConfig(name string, decode decoderFunc) (any, error) {
	t := match(name)
	if t.translate == nil {
		return nil, ErrTranslator
	}
	c, err := t.buildConfig(decode)
	return c, err
}

func Translate(name, text, from, to string, config any) (string, error) {
	t := match(name)
	if t.translate == nil {
		return "", ErrTranslator
	}
	return t.translate(text, from, to, config)
}
