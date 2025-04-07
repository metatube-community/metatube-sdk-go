package openai

import (
	"github.com/sashabaranov/go-openai"
	translator "github.com/xjasonlyu/openai-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*OpenAI)(nil)

const (
	defaultModel        = openai.GPT4
	defaultSystemPrompt = `You are a professional translator for adult video content. You must also follow the following rules:
1. Use official translations for actor/actress names when available, otherwise leave them unchanged
2. Do not make up translations for names without official versions
3. Keep all numbers, dates, and measurements in their original format
4. Translate naturally and fluently, avoiding literal/word-for-word translation
5. Do not add any explanations or notes
6. Output only the content of the translation
`
)

type OpenAI struct {
	APIKey string `json:"openai-api-key"`
	APIUrl string `json:"openai-api-url"`
	Model  string `json:"openai-model"`
}

func (oa *OpenAI) Translate(q, source, target string) (result string, err error) {
	opts := []translator.Option{
		translator.WithFrom(source),
		translator.WithModel(defaultModel),
		translator.WithSystemPrompt(defaultSystemPrompt),
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
