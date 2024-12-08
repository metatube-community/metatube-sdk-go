package deepl

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*DeepL)(nil)

const deeplTranslateAPI = "https://api-free.deepl.com/v2/translate"

type DeepL struct {
	APIKey string `json:"deepl-api-key"`
	// AltURL is an optional DeepLX URL. It is only
	// compatible with the /v2/translate API.
	AltURL string `json:"deepl-alt-url"`
}

func (dpl *DeepL) Translate(q, source, target string) (result string, err error) {
	apiURL := deeplTranslateAPI
	if dpl.AltURL != "" {
		apiURL = dpl.AltURL
	}

	var resp *http.Response
	if resp, err = fetch.Post(
		apiURL,
		fetch.WithURLEncodedBody(map[string]string{
			"text":            q,
			"source_lang":     parseToDeeplSupportedLanguage(source),
			"target_lang":     parseToDeeplSupportedLanguage(target),
			"split_sentences": "0", // disable sentence split
		}),
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Authorization", "DeepL-Auth-Key "+dpl.APIKey),
		fetch.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	data := struct {
		Translations []struct {
			DetectedSourceLanguage string `json:"detected_source_language"`
			Text                   string `json:"text"`
		} `json:"translations"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err == nil {
		result = data.Translations[0].Text
	}
	return
}

func parseToDeeplSupportedLanguage(lang string) string {
	lang = strings.ToUpper(lang)
	switch lang {
	case "ZH", "CHS", "ZH-CN", "ZH-HANS", "CHT", "ZH-TW", "ZH-HK", "ZH-HANT":
		return "ZH"
	case "AUTO":
		return ""
	default:
		return lang
	}
}

func init() {
	translate.Register(&DeepL{})
}
