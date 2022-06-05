package translate

import (
	"encoding/json"
	"net/http"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const googleTranslateAPI = "https://translation.googleapis.com/language/translate/v2"

func GoogleTranslate(q, source, target, key string) (result string, err error) {
	var resp *http.Response
	if resp, err = fetch.Post(
		googleTranslateAPI,
		fetch.WithJSONBody(map[string]string{
			"q":      q,
			"source": source,
			"target": target,
			"format": "text",
		}),
		fetch.WithQuery(map[string]string{"key": key}),
		fetch.WithHeader("Content-Type", "application/json"),
	); err != nil {
		return
	}

	data := struct {
		Data struct {
			Translations []struct {
				DetectedSourceLanguage string `json:"detectedSourceLanguage"`
				TranslatedText         string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}
	return data.Data.Translations[0].TranslatedText, nil
}
