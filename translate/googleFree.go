package translate

import (
	"time"

	"github.com/bregydoc/gtranslate"
)

func GoogleFreeTranslate(q, source, target string) (result string, err error) {
	if result, err = gtranslate.TranslateWithParams(q, gtranslate.TranslationParams{
		From:  source,
		To:    target,
		Tries: 1,
		Delay: time.Second * 10,
	}); err != nil {
		result = q
	}
	return
}
