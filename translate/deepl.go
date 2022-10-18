package translate

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const deeplTranslateAPI = "https://api-free.deepl.com/v2/translate"

type deeplData struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func DeeplTranslate(q, source, target, key string) (result string, err error) {
	var resp *http.Response
	if resp, err = fetch.Post(
		deeplTranslateAPI,
		fetch.WithURLEncodedBody(map[string]string{
			"text":            q,
			"source_lang":     parseToDeeplSupportedLanguage(source),
			"target_lang":     parseToDeeplSupportedLanguage(target),
			"split_sentences": "0", // disable stntence split
		}),
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Authorization", "DeepL-Auth-Key "+key),
		fetch.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	data := new(deeplData)
	if err = json.NewDecoder(resp.Body).Decode(data); err == nil {
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
