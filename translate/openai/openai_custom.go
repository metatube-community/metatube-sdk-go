package openai

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

var _ translate.Translator = (*OpenAIX)(nil)

type OpenAIX struct {
	APIKey       string `json:"openai-api-key"`
	BaseURL      string `json:"base-url"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system-prompt"`
}

func (oa *OpenAIX) Translate(q, source, target string) (result string, err error) {
	if oa.BaseURL == "" {
		return "", translate.ErrInvalidConfiguration
	}

	// Prepare the chat message
	systemPrompt := oa.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are a professional translator for adult video content.
Rules:
1. Use official translations for actor/actress names if available, otherwise keep them unchanged
2. Do not invent translations for names without official versions
3. Maintain any numbers, dates, and measurements in their original format
4. Translate naturally and fluently, avoiding word-for-word translation
5. Do not add any explanations or notes
6. Only output the translation`
	}

	userPrompt := fmt.Sprintf("Please translate the following text from %s to %s:\n\n%s", source, target, q)

	model := oa.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// 准备请求体
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  1000,
	}

	// 准备请求选项
	opts := []fetch.Option{
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Content-Type", "application/json"),
		fetch.WithHeader("Accept", "application/json"),
	}

	// 如果配置了 API Key，添加认证头
	if oa.APIKey != "" {
		opts = append(opts,
			fetch.WithHeader("Authorization", "Bearer "+oa.APIKey),
		)
	}

	var resp *http.Response
	if resp, err = fetch.Post(oa.BaseURL, fetch.WithJSONBody(reqBody), opts...); err != nil {
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
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error,omitempty"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err = json.NewDecoder(bytes.NewReader(respBody)).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode response: %v, body: %s", err, string(respBody))
	}

	if data.Error != nil {
		return "", fmt.Errorf("API error: %s (%s)", data.Error.Message, data.Error.Type)
	}

	if len(data.Choices) > 0 {
		result = strings.TrimSpace(data.Choices[0].Message.Content)
	} else {
		err = fmt.Errorf("no translation result")
	}
	return
}

func init() {
	translate.Register(&OpenAIX{})
}
