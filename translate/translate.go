package translate

import "errors"

type Engine uint8

const (
	Google Engine = iota + 1
)

type Translator interface {
	Translate(text string, srcLang, dstLang string) (string, error)
}

func SetEngine(e Engine) {
	switch e {
	case Google:
		defaultTranslate = NewGoogleTranslator()
	default:
		panic(errors.New("unsupported translate engine"))
	}
}

// Translate translates text from source language to target language.
func Translate(text string, srcLang, dstLang string) (string, error) {
	return defaultTranslate.Translate(text, srcLang, dstLang)
}

var defaultTranslate Translator

func init() {
	SetEngine(Google) // default translator
}
