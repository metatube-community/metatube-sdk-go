package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*OpenAIX)(nil)

type OpenAIX struct {
	APIKey   string `json:"openai-api-key"`
	BaseURL  string `json:"base-url"`
	Model    string `json:"model"`
}

func (oa *OpenAIX) Translate(q, source, target string) (result string, err error) {
	if oa.BaseURL == "" {
		return "", translate.ErrInvalidConfiguration
	}

	// Prepare the chat message
	prompt := fmt.Sprintf(`You are a professional translator for adult video content. Please translate the following text from %s to %s.
Rules:
1. Keep actor/actress names and video codes unchanged
2. Maintain any numbers, dates, and measurements in their original format
3. Translate naturally and fluently, avoiding word-for-word translation
4. Do not add any explanations or notes
5. Only output the translation

Text to translate:
%s`, source, target, q)
	
	model := oa.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	var resp *http.Response
	if resp, err = fetch.Post(
		oa.BaseURL,
		fetch.WithJSONBody(bytes.NewReader(reqJSON)),
		fetch.WithRaiseForStatus(true),
		fetch.WithHeader("Authorization", "Bearer "+oa.APIKey),
		fetch.WithHeader("Content-Type", "application/json"),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	var data struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	if len(data.Choices) > 0 {
		result = strings.TrimSpace(data.Choices[0].Message.Content)
	}
	return
}

func init() {
	translate.Register(&OpenAIX{})
}
