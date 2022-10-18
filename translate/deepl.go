package translate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const deelpTranslateAPI = "https://api-free.deepl.com/v2/translate"

type deeplData struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func DeeplTranslate(q, source, target, key string) (result string, err error) {
	var resp *http.Response
	if resp, err = fetch.Post(
		deelpTranslateAPI,
		fetch.WithURLEncodedBody(map[string]string{
			"text":            q,
			"source_lang":     parseToDeeplSupportedLanguage(source),
			"target_lang":     parseToDeeplSupportedLanguage(target),
			"split_sentences": "0", // disable stntence split
		}),
		fetch.WithRaiseForStatus(false),
		fetch.WithHeader("Authorization", "DeepL-Auth-Key "+key),
		fetch.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	data := new(deeplData)
	if err = json.NewDecoder(resp.Body).Decode(data); err == nil {
		if len(data.Translations) == 0 {
			err = fmt.Errorf("translate text is nil: %v", data.Translations)
			return
		}
		result = data.Translations[0].Text
	}
	return
}

func parseToDeeplSupportedLanguage(lang string) string {
	lang = strings.ToLower(lang)
	switch lang {
	case "zh", "chs", "zh-cn", "zh_cn", "zh-hans":
		return "ZH"
	case "cht", "zh-tw", "zh_tw", "zh-hk", "zh_hk", "zh-hant":
		return "ZH"
	case "jp", "ja":
		return "JA"
	case "kor", "ko":
		return ""
	case "vie", "vi":
		return ""
	case "spa", "es":
		return "EL"
	case "de":
		return "DE"
	case "fra", "fr":
		return "FR"
	case "ara", "ar":
		return ""
	case "bul", "bg":
		return "BG"
	case "est", "et":
		return "ET"
	case "dan", "da":
		return "DA"
	case "fin", "fi":
		return ""
	case "rom", "ro":
		return "RO"
	case "slo", "sl":
		return "SL"
	case "swe", "sv":
		return "SV"
	default:
		return ""
	}
}
