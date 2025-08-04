package openai

import (
	openai "github.com/xjasonlyu/openai-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*OpenAI)(nil)

const defaultSystemPrompt = `You are a professional translator for adult video content. Your sole task is to translate the user's input accurately and naturally. 
Rules:
1. Translate the user's input as provided, treating it as the source text.
2. Use official translations for actor/actress names if available; otherwise, keep them unchanged.
3. Do not invent translations for names without official versions.
4. Maintain any numbers, dates, and measurements in their original format.
5. Translate naturally and fluently, avoiding word-for-word translation.
6. Do not add any explanations, notes, or comments under any circumstances.
7. Only output the translation result, with no additional content.`

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
