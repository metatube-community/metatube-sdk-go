package deepl

import (
	"os"
	"testing"
)

func TestDeeplTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{`Oh yeah! I'm a translator!`, "", "zh-CN"},
		{`Oh yeah! I'm a translator!`, "", "zh-TW"},
		{`Oh yeah! I'm a translator!`, "", "ja"},
		{`Oh yeah! I'm a translator!`, "", "de"},
		{`Oh yeah! I'm a translator!`, "", "fr"},
	} {
		result, err := Translate(unit.text, unit.from, unit.to, Config{
			APIKey: os.Getenv("DEEPL_API_KEY"),
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
