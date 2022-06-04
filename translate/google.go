package translate

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const googleTranslateAPI = "https://translate.google.com/translate_a/single"

type GoogleTranslator struct {
	fetcher *fetch.Fetcher
}

func NewGoogleTranslator() Translator {
	return &GoogleTranslator{
		fetcher: fetch.Default(&fetch.Config{
			RandomUserAgent: true,
		}),
	}
}

func (gt *GoogleTranslator) Translate(text, srcLang, dstLang string) (result string, err error) {
	resp, err := gt.fetcher.Get(googleTranslateAPI,
		fetch.WithQuery(map[string]string{
			"client": "at",
			"dt":     "t",
			"dj":     "1",
			"ie":     "UTF-8",
			"oe":     "UTF-8",
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

var _ Translator = (*GoogleTranslator)(nil)
