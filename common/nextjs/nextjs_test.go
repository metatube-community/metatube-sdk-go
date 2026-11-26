package nextjs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const samplePage = `<html><head>
<script src="/_next/static/chunks/main.js" async=""></script>
</head><body>
<script>document.write("not a flight data script")</script>
<script>(self.__next_f=self.__next_f||[]).push([0])</script>
<script>self.__next_f.push([1,"1:\"$Sreact.fragment\"\n2:I[123,[],\"default\"]\n"])</script>
<script>self.__next_f.push([1,"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\","])</script>
<script>self.__next_f.push([1,"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n"])</script>
<script>self.__next_f.push([2,"ignored"])</script>
</body></html>`

const samplePayload = "1:\"$Sreact.fragment\"\n" +
	"2:I[123,[],\"default\"]\n" +
	"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\"," +
	"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n"

func TestExtractPayload(t *testing.T) {
	for _, unit := range []struct {
		page    string
		want    string
		wantErr bool
	}{
		{page: samplePage, want: samplePayload},
		{page: `<html></html>`, want: ""},
		{page: `<script>self.__next_f.push(</script>`, wantErr: true},
	} {
		payload, err := ExtractPayload([]byte(unit.page))
		if unit.wantErr {
			assert.Error(t, err, unit.page)
			continue
		}
		require.NoError(t, err, unit.page)
		assert.Equal(t, unit.want, payload)
	}
}

func TestExtractPayloadTimeout(t *testing.T) {
	defer func(d time.Duration) { scriptTimeout = d }(scriptTimeout)
	scriptTimeout = 100 * time.Millisecond

	_, err := ExtractPayload([]byte(`<script>self.__next_f; while (true) {}</script>`))
	assert.ErrorIs(t, err, errScriptTimeout)
}

func TestFindObject(t *testing.T) {
	for _, unit := range []struct {
		keys []string
		want map[string]string // fields of the found object; nil if not found.
	}{
		{[]string{"src", "poster"}, map[string]string{
			"src":      "https://example.com/a.m3u8",
			"poster":   "https://example.com/a.jpg",
			"movieSeq": "1",
		}},
		{[]string{"src", "missing"}, nil},
		{nil, nil},
	} {
		var got map[string]any
		if !assert.Equal(t, unit.want != nil, FindObject(samplePayload, &got, unit.keys...), unit.keys) {
			continue
		}
		for key, want := range unit.want {
			assert.Equal(t, want, fmt.Sprint(got[key]), key)
		}
	}
}
