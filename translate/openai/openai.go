package openai

import (
	translator "github.com/xjasonlyu/openai-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*OpenAI)(nil)

const defaultSystemPrompt = `You are a professional translator for adult video content.
Rules:
1. Use official translations for actor/actress names if available, otherwise keep them unchanged
2. Do not invent translations for names without official versions
3. Maintain any numbers, dates, and measurements in their original format
4. Translate naturally and fluently, avoiding word-for-word translation
5. Do not add any explanations or notes
6. Only output the translation`

type OpenAI struct {
	APIKey string `json:"openai-api-key"`
	APIUrl string `json:"openai-api-url"`
	Model  string `json:"openai-model"`
}

func (oa *OpenAI) Translate(q, source, target string) (result string, err error) {
	opts := []translator.Option{
		translator.WithSourceLanguage(source),
		translator.WithSystemPrompt(defaultSystemPrompt),
	}
	if oa.APIUrl != "" {
		opts = append(opts, translator.WithBaseURL(oa.APIUrl))
	}
	if oa.Model != "" {
		opts = append(opts, translator.WithModel(oa.Model)) // overwrite
	}
	return translator.Translate(q, target, oa.APIKey, opts...)
}

func init() {
	translate.Register(&OpenAI{})
}
