package deeplx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	// 如果源语言为空，设置为 auto
	if source == "" {
		source = "auto"
	}

	reqBody := map[string]string{
		"text":        q,
		"source_lang": parseToDeeplSupportedLanguage(source),
		"target_lang": parseToDeeplSupportedLanguage(target),
	}

	// 准备请求选项
	opts := []fetch.Option{
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Content-Type", "application/json"),
		fetch.WithHeader("Accept", "*/*"),
		fetch.WithHeader("User-Agent", "MetaTube/1.0.0"),
		fetch.WithHeader("Connection", "keep-alive"),
	}

	// 如果配置了 API Key，添加认证头
	if dpl.APIKey != "" {
		opts = append(opts,
			fetch.WithHeader("Authorization", "DeepL-Auth-Key "+dpl.APIKey),
		)
	}

	var resp *http.Response
	if resp, err = fetch.Post(dpl.BaseURL, fetch.WithJSONBody(reqBody), opts...); err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	fmt.Printf("Response: %s\n", string(respBody))

	var data struct {
		Code   int    `json:"code"`
		Data   string `json:"data"`
		Method string `json:"method"`
	}
	if err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode response: %v, body: %s", err, string(respBody))
	}

	if data.Code == 200 {
		result = data.Data
	} else {
		err = fmt.Errorf("translation failed with code %d: %s", data.Code, data.Data)
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
