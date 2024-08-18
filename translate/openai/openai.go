package openai

import (
	openai "github.com/zijiren233/openai-translator"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

type Config struct {
	APIKey string `json:"openai-api-key"`
}

func Translate(q, source, target string, config Config) (result string, err error) {
	return openai.Translate(q, target, config.APIKey, openai.WithFrom(source))
}

func init() {
	translate.Register("openai", Translate, func() Config {
		return Config{}
	})
}
