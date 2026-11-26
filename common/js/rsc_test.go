package js

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

func TestExtractRSCPayload(t *testing.T) {
	payload, err := ExtractRSCPayload([]byte(samplePage))
	require.NoError(t, err)
	assert.Equal(t, samplePayload, payload)

	payload, err = ExtractRSCPayload([]byte(`<html></html>`))
	require.NoError(t, err)
	assert.Empty(t, payload)

	_, err = ExtractRSCPayload([]byte(`<script>self.__next_f.push(</script>`))
	assert.Error(t, err)

	// Binary chunks are base64-decoded.
	payload, err = ExtractRSCPayload([]byte(`<script>self.__next_f.push([1,"0:[\"a"])</script>` +
		`<script>self.__next_f.push([3,"YmM="])</script>` +
		`<script>self.__next_f.push([1,"\"]\n"])</script>`))
	require.NoError(t, err)
	assert.Equal(t, "0:[\"abc\"]\n", payload)
}

func TestExtractRSCPayloadTimeout(t *testing.T) {
	defer func(d time.Duration) { rscScriptTimeout = d }(rscScriptTimeout)
	rscScriptTimeout = 100 * time.Millisecond

	_, err := ExtractRSCPayload([]byte(`<script>self.__next_f; while (true) {}</script>`))
	assert.ErrorIs(t, err, errRSCScriptTimeout)
}

func TestFindRSCObject(t *testing.T) {
	type player struct {
		Src      string `json:"src"`
		Poster   string `json:"poster"`
		MovieSeq int    // untagged, matched case-insensitively.
		Title    string `json:"title"` // not in the payloads.
	}

	find := func(payload string, keys ...string) (p player) {
		require.True(t, FindRSCObject(payload, &p, keys...))
		return
	}

	assert.Equal(t, player{
		Src:      "https://example.com/a.m3u8",
		Poster:   "https://example.com/a.jpg",
		MovieSeq: 1,
	}, find(samplePayload))

	// Objects without any field of the struct never match, even with keys.
	var unknown struct{ Foo string }
	assert.False(t, FindRSCObject(samplePayload, &unknown))
	assert.False(t, FindRSCObject(samplePayload, &unknown, "src"))

	// The second object matches more fields than the first one.
	const payload = `0:[{"src":"a.jpg","alt":"a"},{"src":"b.m3u8","poster":"b.jpg","movieseq":2}]` + "\n"
	assert.Equal(t, player{Src: "b.m3u8", Poster: "b.jpg", MovieSeq: 2}, find(payload))

	// Keys narrow down the candidates.
	assert.Equal(t, player{Src: "a.jpg"}, find(payload, "alt"))
	assert.False(t, FindRSCObject(payload, new(player), "missing"))

	// A map has no fields to match, so keys are required.
	var m map[string]any
	assert.False(t, FindRSCObject(payload, &m))
	require.True(t, FindRSCObject(payload, &m, "poster"))
	assert.Equal(t, "b.m3u8", m["src"])

	// Length-prefixed rows are skipped by their length, even if they contain
	// newlines.
	const text = "a\n7:{\"src\":\"fake\"}\n"
	assert.Equal(t, player{Src: "real"},
		find(fmt.Sprintf("5:T%x,%s6:{\"src\":\"real\"}\n", len(text), text)))
	assert.Equal(t, player{Src: "real"}, find("5:A4,\x00\x01\x02\x036:{\"src\":\"real\"}\n"))

	// Keys that differ only in case count once.
	assert.Equal(t, player{Src: "a", Poster: "b"},
		find(`0:{"src":"x","SRC":"y"}`+"\n"+`1:{"src":"a","poster":"b"}`+"\n"))

	// v is left untouched if the best object cannot be unmarshaled into it,
	// e.g., a field holding a reference to another row.
	var tagged struct {
		Src  string   `json:"src"`
		Tags []string `json:"tags"`
	}
	tagged.Src = "untouched"
	assert.False(t, FindRSCObject(`0:{"src":"a","tags":"$4"}`+"\n", &tagged))
	assert.Equal(t, "untouched", tagged.Src)
	assert.Nil(t, tagged.Tags)

	// The object is merged into v, keeping defaults it does not have.
	merged := player{Title: "default"}
	require.True(t, FindRSCObject(`0:{"src":"a","poster":"b"}`+"\n", &merged))
	assert.Equal(t, player{Src: "a", Poster: "b", Title: "default"}, merged)

	kept := map[string]any{"keep": 1}
	require.True(t, FindRSCObject(`0:{"src":"a"}`+"\n", &kept, "src"))
	assert.Equal(t, map[string]any{"keep": 1, "src": "a"}, kept)
}

func TestFindRSCObjectDecodeValues(t *testing.T) {
	var v struct {
		Src   string    `json:"src"`
		Price string    `json:"price"`
		Date  time.Time `json:"date"`
		Size  int64     `json:"size"`
		NaN   *float64  `json:"nan"`
	}
	require.True(t, FindRSCObject(`0:{"src":"$undefined","price":"$$5",`+
		`"date":"$D2026-11-23T00:00:00.000Z","size":"$n123","nan":"$NaN"}`+"\n", &v))
	assert.Empty(t, v.Src)
	assert.Equal(t, "$5", v.Price)
	assert.Equal(t, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC), v.Date)
	assert.Equal(t, int64(123), v.Size)
	assert.Nil(t, v.NaN)
}
