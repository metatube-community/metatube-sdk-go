package translate

import "testing"

func TestTranslate(t *testing.T) {
	for _, engine := range []Engine{
		Google,
	} {
		SetEngine(engine)
		result, err := Translate(`Oh yeah! I'm a translator!`, "auto", "zh_cn")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
