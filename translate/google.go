package translate

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/text/language"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const googleTranslateAPI = "https://translation.googleapis.com/language/translate/v2"

func GoogleTranslate(q, source, target, key string) (result string, err error) {
	var resp *http.Response
	if resp, err = fetch.Post(
		googleTranslateAPI,
		fetch.WithJSONBody(map[string]string{
			"q":      q,
			"source": parseToGoogleSupportedLanguage(source),
			"target": parseToGoogleSupportedLanguage(target),
			"format": "text",
		}),
		fetch.WithRaiseForStatus(false),
		fetch.WithQuery("key", key),
		fetch.WithHeader("Content-Type", "application/json"),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	data := struct {
		Data struct {
			Translations []struct {
				DetectedSourceLanguage string `json:"detectedSourceLanguage"`
				TranslatedText         string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err == nil {
		if len(data.Data.Translations) > 0 {
			result = data.Data.Translations[0].TranslatedText
		} else if data.Error.Code > 0 {
			err = errors.New(data.Error.Message)
		} else {
			err = errors.New("google translate: unknown error")
		}
	}
	return
}

func parseToGoogleSupportedLanguage(lang string) string {
	if lang = strings.ToLower(lang); lang == "" || lang == "auto" /* auto detect */ {
		return ""
	}
	tag, err := language.Parse(lang)
	if err != nil {
		return lang /* fallback to original */
	}
	return tag.String()
}
