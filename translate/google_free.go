package translate

import (
	"time"

	translater "github.com/zijiren233/google-translater"
)

func GoogleFreeTranslate(q, source, target string) (string, error) {
	data, err := translater.Translate(q, translater.TranslationParams{
		From:       source,
		To:         target,
		Retry:      2,
		RetryDelay: 3 * time.Second,
	})
	return data.Text, err
}
