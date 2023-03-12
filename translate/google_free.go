package translate

import (
	"math/rand"
	"time"

	translater "github.com/zijiren233/google-translater"
)

func GoogleFreeTranslate(q, source, target string) (string, error) {
	data, err := translater.Translate(q, target, translater.TranslationParams{
		From:       source,
		Retry:      2,
		RetryDelay: 3 * time.Second,
	})
	if err != nil {
		data, err = translater.TranslateWithClienID(q, target, translater.TranslationWithClienIDParams{
			From:       source,
			Retry:      2,
			ClientID:   rand.Intn(5) + 1,
			RetryDelay: 3 * time.Second,
		})
	}
	return data.Text, err
}
