package nextjs

import (
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
	payload, err := ExtractPayload([]byte(samplePage))
	require.NoError(t, err)
	assert.Equal(t, samplePayload, payload)

	payload, err = ExtractPayload([]byte(`<html></html>`))
	require.NoError(t, err)
	assert.Empty(t, payload)

	_, err = ExtractPayload([]byte(`<script>self.__next_f.push(</script>`))
	assert.Error(t, err)
}

func TestExtractPayloadTimeout(t *testing.T) {
	defer func(d time.Duration) { scriptTimeout = d }(scriptTimeout)
	scriptTimeout = 100 * time.Millisecond

	_, err := ExtractPayload([]byte(`<script>self.__next_f; while (true) {}</script>`))
	assert.ErrorIs(t, err, errScriptTimeout)
}

func TestFindObject(t *testing.T) {
	type player struct {
		Src      string `json:"src"`
		Poster   string `json:"poster"`
		MovieSeq int    // untagged, matched case-insensitively.
		Title    string `json:"title"` // not in the payloads.
	}

	var p player
	require.True(t, FindObject(samplePayload, &p))
	assert.Equal(t, player{
		Src:      "https://example.com/a.m3u8",
		Poster:   "https://example.com/a.jpg",
		MovieSeq: 1,
	}, p)

	var unknown struct{ Foo string }
	assert.False(t, FindObject(samplePayload, &unknown))

	// The second object matches more fields than the first one.
	const payload = `0:[{"src":"a.jpg","alt":"a"},{"src":"b.m3u8","poster":"b.jpg","movieseq":2}]` + "\n"

	p = player{}
	require.True(t, FindObject(payload, &p))
	assert.Equal(t, player{Src: "b.m3u8", Poster: "b.jpg", MovieSeq: 2}, p)

	// Keys narrow down the candidates.
	p = player{}
	require.True(t, FindObject(payload, &p, "alt"))
	assert.Equal(t, player{Src: "a.jpg"}, p)
	assert.False(t, FindObject(payload, &p, "missing"))

	// A map has no fields to match, so keys are required.
	var m map[string]any
	assert.False(t, FindObject(payload, &m))
	require.True(t, FindObject(payload, &m, "poster"))
	assert.Equal(t, "b.m3u8", m["src"])
}
