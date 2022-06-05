package translate

import (
	"os"
	"testing"
)

func TestGoogleTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{`Oh yeah! I'm a translator!`, "", "zh-CN"},
		{`Oh yeah! I'm a translator!`, "", "zh-TW"},
		{`Oh yeah! I'm a translator!`, "", "ja"},
		{`Oh yeah! I'm a translator!`, "", "de"},
		{`Oh yeah! I'm a translator!`, "", "fr"},
	} {
		result, err := GoogleTranslate(unit.text, unit.from, unit.to, os.Getenv("GOOGLE_API_KEY"))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
