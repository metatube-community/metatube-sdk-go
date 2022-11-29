package translate

import (
	"time"

	translater "github.com/zijiren233/google-translater"
)

func GoogleFreeTranslate(q, source, target string) (result string, err error) {
	if data, err := translater.Translate(q, translater.TranslationParams{
		From:       source,
		To:         target,
		Retry:      2,
		RetryDelay: time.Second * 3,
	}); err != nil {
		result = q
	} else {
		result = data.Text
	}
	return
}
