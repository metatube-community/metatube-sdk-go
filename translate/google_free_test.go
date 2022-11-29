package translate

import (
	"testing"
)

func TestGoogleFreeTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{"Oh yeah!\nI'm a translator!", "", "zh-CN"},
		{"Oh yeah!\nI'm a translator!", "", "zh-TW"},
		{"Oh yeah!\nI'm a translator!", "", "ja"},
		{"Oh yeah!\nI'm a translator!", "", "de"},
		{"Oh yeah!\nI'm a translator!", "", "fr"},
	} {
		result, err := GoogleFreeTranslate(unit.text, unit.from, unit.to)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
