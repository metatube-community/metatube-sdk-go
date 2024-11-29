package deeplx

import (
	"os"
	"testing"
)

func TestDeepLXTranslate(t *testing.T) {
	baseURL := os.Getenv("DEEPLX_BASE_URL")
	if baseURL == "" {
		t.Skip("DEEPLX_BASE_URL not set")
	}

	apiKey := os.Getenv("DEEPLX_API_KEY")

	translator := &DeepLX{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}

	for _, unit := range []struct {
		text, from, to string
	}{
		{`Oh yeah! I'm a translator!`, "auto", "ZH"},
		{`私は翻訳者です`, "JA", "ZH"},
		{`我是一个翻译器`, "ZH", "EN"},
		{`PPPE-001 深田えいみ 親友からこっそり彼氏を寝取る巨乳でエッチな痴女お姉さん`, "JA", "ZH"},
		{`A Busty and Naughty Sister Who Secretly Takes Her Best Friend's Boyfriend`, "EN", "ZH"},
	} {
		result, err := translator.Translate(unit.text, unit.from, unit.to)
		if err != nil {
			t.Errorf("Failed to translate text %q from %q to %q: %v", unit.text, unit.from, unit.to, err)
			continue
		}
		t.Logf("Translated %q (%s->%s): %q", unit.text, unit.from, unit.to, result)
	}
}
