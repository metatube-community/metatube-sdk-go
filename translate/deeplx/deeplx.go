package deeplx

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*DeepLX)(nil)

type DeepLX struct {
	APIKey  string `json:"deepl-api-key"`
	BaseURL string `json:"base-url"`
}

func (dpl *DeepLX) Translate(q, source, target string) (result string, err error) {
	if dpl.BaseURL == "" {
		return "", translate.ErrInvalidConfiguration
	}

	var resp *http.Response
	if resp, err = fetch.Post(
		dpl.BaseURL,
		fetch.WithURLEncodedBody(map[string]string{
			"text":            q,
			"source_lang":     parseToDeeplSupportedLanguage(source),
			"target_lang":     parseToDeeplSupportedLanguage(target),
			"split_sentences": "0", // disable sentence split
		}),
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Authorization", "DeepL-Auth-Key "+dpl.APIKey),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	var data struct {
		Code   int    `json:"code"`
		Data   string `json:"data"`
		Method string `json:"method"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	if data.Code == 200 {
		result = data.Data
	}
	return
}

func parseToDeeplSupportedLanguage(lang string) string {
	switch strings.ToLower(lang) {
	case "zh", "zh-hans", "zh-cn", "chs":
		return "ZH"
	case "zh-hant", "zh-tw", "cht":
		return "ZH"
	case "en":
		return "EN"
	case "ja", "jp":
		return "JA"
	default:
		return strings.ToUpper(lang)
	}
}

func init() {
	translate.Register(&DeepLX{})
}
