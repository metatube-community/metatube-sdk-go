package translate

import (
	"math/rand"
	"time"

	translator "github.com/zijiren233/google-translater" //nolint:misspell
)

func GoogleFreeTranslate(q, source, target string) (string, error) {
	data, err := translator.Translate(q, target, translator.TranslationParams{
		From:       source,
		Retry:      2,
		RetryDelay: 3 * time.Second,
	})
	if err != nil {
		data, err = translator.TranslateWithClienID(q, target, translator.TranslationWithClienIDParams{
			From:       source,
			Retry:      2,
			ClientID:   rand.Intn(5) + 1,
			RetryDelay: 3 * time.Second,
		})
	}
	return data.Text, err
}
