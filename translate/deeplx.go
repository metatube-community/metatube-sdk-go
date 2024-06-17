package translate

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
)

const deeplxTranslateAPITemplate = "https://api.deeplx.org/{api-key}/translate"

func buildDeeplxTranslateURL(apiKey string) string {
	return strings.Replace(deeplxTranslateAPITemplate, "{api-key}", apiKey, 1)
}

func DeepLXTranslate(q, source, target, apiKey string) (result string, err error) {
	url := buildDeeplxTranslateURL(apiKey)
	var resp *http.Response
	if resp, err = fetch.Post(
		url,
		fetch.WithJSONBody(map[string]string{
			"text":        q,
			"source_lang": parseToDeeplxSupportedLanguage(source),
			"target_lang": parseToDeeplxSupportedLanguage(target),
		}),
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Content-Type", "application/json"),
	); err != nil {
		return
	}
	defer resp.Body.Close()
	data := struct {
		Data string `json:"data"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err == nil {
		result = data.Data
	}
	return
}

func parseToDeeplxSupportedLanguage(lang string) string {
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
