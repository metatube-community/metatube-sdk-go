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
		fetch.WithJSONBody(map[string]any{
			"text":            splitTextsAfter(q, "\n", ".", "。"),
			"source_lang":     parseToDeeplSupportedLanguage(source),
			"target_lang":     parseToDeeplSupportedLanguage(target),
			"split_sentences": "0", // disable sentence split
		}),
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Authorization", "DeepL-Auth-Key "+dpl.APIKey),
		fetch.WithHeader("Content-Type", "application/json"),
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
		sb := &strings.Builder{}
		for _, tl := range data.Translations {
			sb.WriteString(tl.Text)
		}
		result = sb.String()
	}
	return
}

func splitTextsAfter(text string, seps ...string) []string {
	results := []string{text}
	for _, sep := range seps {
		var temp []string
		for _, str := range results {
			parts := strings.SplitAfter(str, sep)
			temp = append(temp, parts...)
		}
		results = temp
	}
	return results
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
