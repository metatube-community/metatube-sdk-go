package translate

import (
	openai "github.com/zijiren233/openai-translator"
)

func OpenaiTranslate(q, source, target, key string) (result string, err error) {
	return openai.Translate(q, target, key, openai.WithFrom(source))
}
