package deepl

import (
	"strings"

	deeplx "github.com/xjasonlyu/deeplx-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*DeepL)(nil)

type DeepL struct {
	APIKey string `json:"deepl-api-key"`
	// APIUrl is an optional DeepL(X) URL.
	APIUrl string `json:"deepl-api-url"`
}

func (dpl *DeepL) Translate(q, source, target string) (result string, err error) {
	var opts []deeplx.TranslatorOption
	if dpl.APIUrl != "" {
		opts = append(opts, deeplx.WithBaseURL(dpl.APIUrl))
	}
	return deeplx.
		NewTranslator(dpl.APIKey, opts...).
		TranslateText(q,
			parseToSupportedLanguage(target),
			deeplx.WithSourceLang(
				parseToSupportedLanguage(source)),
		)
}

func parseToSupportedLanguage(lang string) string {
	lang = strings.ToUpper(lang)
	switch lang {
	case "CHS", "ZH-CN", "ZH-HANS":
		return "ZH"
	case "CHT", "ZH-TW", "ZH-HK":
		return "ZH-HANT"
	case "AUTO":
		return ""
	default:
		return lang
	}
}

func init() {
	translate.Register(&DeepL{})
}
