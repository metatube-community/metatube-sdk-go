package openai

import (
	"github.com/sashabaranov/go-openai"
	translator "github.com/zijiren233/openai-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*OpenAI)(nil)

const defaultModel = openai.GPT4o

type OpenAI struct {
	APIKey string `json:"openai-api-key"`
	APIUrl string `json:"openai-api-url"`
	Model  string `json:"openai-model"`
}

func (oa *OpenAI) Translate(q, source, target string) (result string, err error) {
	opts := []translator.Option{
		translator.WithFrom(source),
		translator.WithModel(defaultModel),
	}
	if oa.APIUrl != "" {
		opts = append(opts, translator.WithUrl(oa.APIUrl))
	}
	if oa.Model != "" {
		opts = append(opts, translator.WithModel(oa.Model)) // overwrite
	}
	return translator.Translate(q, target, oa.APIKey, opts...)
}

func init() {
	translate.Register(&OpenAI{})
}
