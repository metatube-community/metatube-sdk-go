package translate

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"go.uber.org/atomic"
)

var ErrTranslator = &errorTranslator{errors.New("translate: unknown translator")}

type Translator interface {
	Translate(text, from, to string) (string, error)
}

var (
	_ error      = (*errorTranslator)(nil)
	_ Translator = (*errorTranslator)(nil)
)

type errorTranslator struct{ error }

func (e errorTranslator) Translate(string, string, string) (string, error) {
	return "", e.error
}

type factory struct {
	name string
	new  func() Translator
}

// Translators is the list of registered translators.
var (
	translatorsMu     sync.Mutex
	atomicTranslators atomic.Value
)

// Register registers a translator to package.
func Register(translator Translator) {
	translatorsMu.Lock()
	translators, _ := atomicTranslators.Load().([]factory)
	atomicTranslators.Store(append(translators, factory{
		name: strings.ToLower(
			reflect.TypeOf(translator).Elem().Name()),
		new: func() Translator {
			return reflect.New(reflect.TypeOf(translator).
				Elem()).Interface().(Translator)
		},
	}))
	translatorsMu.Unlock()
}

func match(name string) factory {
	translators := atomicTranslators.Load().([]factory)
	for _, t := range translators {
		if t.name == name {
			return t
		}
	}
	return factory{}
}

func New(name string, unmarshal func(any) error) Translator {
	f := match(name)
	if f.new == nil {
		return ErrTranslator
	}
	t := f.new()
	if err := unmarshal(t); err != nil {
		return &errorTranslator{err}
	}
	return t
}
