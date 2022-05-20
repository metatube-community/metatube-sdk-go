package translate

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const googleTranslateAPI = "https://translate.google.com/translate_a/single"

type GoogleTranslator struct{}

func NewGoogleTranslator() Translator { return new(GoogleTranslator) }

func (gt *GoogleTranslator) Translate(text, srcLang, dstLang string) (result string, err error) {
	resp, err := fetch.Fetch(googleTranslateAPI, fetch.WithQuery(map[string]string{
		"client": "gtx",
		"dt":     "t",
		"dj":     "1",
		"ie":     "utf-8",
		"sl":     srcLang,
		"tl":     dstLang,
		"q":      text,
	}))
	if err != nil {
		return
	}
	data := struct {
		Sentences []struct {
			Trans string `json:"trans"`
			Orig  string `json:"orig"`
		} `json:"sentences"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err == nil {
		if len(data.Sentences) == 0 {
			err = errors.New("bad translation")
			return
		}
		s := strings.Builder{}
		for _, sentence := range data.Sentences {
			s.WriteString(sentence.Trans)
		}
		result = s.String()
	}
	return
}
