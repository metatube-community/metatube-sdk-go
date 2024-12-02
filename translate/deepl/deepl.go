package deepl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*DeepL)(nil)

const deeplTranslateAPI = "https://api-free.deepl.com/v2/translate"

type DeepL struct {
	APIKey  string `json:"deepl-api-key"`
	BaseURL string `json:"deepl-base-url"`
}

func (dpl *DeepL) Translate(q, source, target string) (result string, err error) {
	if source == "" {
		source = "auto"
	}

	apiURL := deeplTranslateAPI
	if dpl.BaseURL != "" {
		apiURL = dpl.BaseURL
	}

	reqBody := map[string]string{
		"text":            q,
		"source_lang":     parseToDeeplSupportedLanguage(source),
		"target_lang":     parseToDeeplSupportedLanguage(target),
		"split_sentences": "0", // disable sentence split
	}

	opts := []fetch.Option{
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Content-Type", "application/json"),
	}

	if dpl.APIKey != "" {
		if strings.Contains(apiURL, "/v2/") {
			opts = append(opts,
				fetch.WithHeader("Authorization", "DeepL-Auth-Key "+dpl.APIKey),
			)
		} else {
			opts = append(opts,
				fetch.WithHeader("Authorization", "Bearer "+dpl.APIKey),
			)
		}
	}

	var resp *http.Response
	if resp, err = fetch.Post(
		apiURL,
		fetch.WithJSONBody(reqBody),
		opts...,
	); err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Try to parse as DeepL Official API response first
	data := struct {
		Translations []struct {
			DetectedSourceLanguage string `json:"detected_source_language"`
			Text                   string `json:"text"`
		} `json:"translations"`
	}{}
	if err = json.Unmarshal(respBody, &data); err == nil && len(data.Translations) > 0 {
		result = data.Translations[0].Text
		return
	}

	// Try to parse as DeepLX response
	var deeplxData struct {
		Code   int    `json:"code"`
		Data   string `json:"data"`
		Method string `json:"method"`
	}
	if err = json.Unmarshal(respBody, &deeplxData); err != nil {
		return "", fmt.Errorf("failed to decode response: %v, body: %s", err, string(respBody))
	}

	if deeplxData.Code == 200 {
		result = deeplxData.Data
	} else {
		err = fmt.Errorf("translation failed with code %d: %s", deeplxData.Code, deeplxData.Data)
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
