package openai

import (
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestOpenaiTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{`Oh yeah! I'm a translator!`, "", "zh-CN"},
		{`Oh yeah! I'm a translator!`, "", "zh-TW"},
		{`Oh yeah! I'm a translator!`, "", "ja"},
		{`Oh yeah! I'm a translator!`, "", "de"},
		{`Oh yeah! I'm a translator!`, "", "fr"},
	} {
		result, err := (&OpenAI{
			APIKey: os.Getenv("OPENAI_API_KEY"),
			APIUrl: os.Getenv("OPENAI_API_URL"),
			Model:  openai.GPT4o,
		}).Translate(unit.text, unit.from, unit.to)
		if assert.NoError(t, err) {
			t.Log(result)
		}
	}
}
