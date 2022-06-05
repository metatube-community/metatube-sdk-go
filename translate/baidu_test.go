package translate

import (
	"os"
	"testing"
)

func TestBaiduTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{`Oh yeah! I'm a translator!`, "auto", "zh-CN"},
		{`Oh yeah! I'm a translator!`, "auto", "zh-TW"},
		{`Oh yeah! I'm a translator!`, "auto", "ja"},
		{`Oh yeah! I'm a translator!`, "auto", "de"},
		{`Oh yeah! I'm a translator!`, "auto", "fr"},
	} {
		result, err := BaiduTranslate(unit.text, unit.from, unit.to, os.Getenv("BAIDU_APP_ID"), os.Getenv("BAIDU_APP_KEY"))
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
