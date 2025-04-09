package openai

import (
	openai "github.com/xjasonlyu/openai-translator"

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
	Prompt string `json:"openai-prompt"`
}

func (oa *OpenAI) Translate(q, source, target string) (result string, err error) {
	var opts []openai.TranslatorOption
	if oa.APIUrl != "" {
		opts = append(opts, openai.WithBaseURL(oa.APIUrl))
	}
	return openai.
		NewTranslator(oa.APIKey, opts...).
		TranslateText(q, target,
			openai.WithModel(oa.Model),
			openai.WithSourceLanguage(source),
			openai.WithSystemPrompt(map[bool]string{
				true:  oa.Prompt,
				false: defaultSystemPrompt,
			}[oa.Prompt != ""]),
		)
}

func init() {
	translate.Register(&OpenAI{})
}
