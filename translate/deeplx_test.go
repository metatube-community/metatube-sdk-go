package translate

import (
	"os"
	"testing"
)

func TestDeeplxTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{`Oh yeah! I'm a translator!`, "EN", "ZH"},
		{`Oh yeah! I'm a translator!`, "EN", "zh-TW"},
		{`Oh yeah! I'm a translator!`, "EN", "ja"},
		{`Oh yeah! I'm a translator!`, "EN", "de"},
		{`Oh yeah! I'm a translator!`, "EN", "fr"},
	} {
		result, err := DeepLXTranslate(unit.text, unit.from, unit.to, os.Getenv("DEEPLX_API_KEY"))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
