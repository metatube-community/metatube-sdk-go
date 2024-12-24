package openai

import (
	openai "github.com/zijiren233/openai-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*OpenAI)(nil)

type OpenAI struct {
	APIKey string `json:"openai-api-key"`
}

func (oa *OpenAI) Translate(q, source, target string) (result string, err error) {
	return openai.Translate(q, target, oa.APIKey, openai.WithFrom(source))
}

func init() {
	translate.Register(&OpenAI{})
}
