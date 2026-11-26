package rsc

import (
	"context"
	json "encoding/json/v2"
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

func TestExtractNextPayload(t *testing.T) {
	payload, err := ExtractNextPayload([]byte(samplePage))
	require.NoError(t, err)
	assert.Equal(t, samplePayload, payload)

	payload, err = ExtractNextPayload([]byte(`<html></html>`))
	require.NoError(t, err)
	assert.Empty(t, payload)

	_, err = ExtractNextPayload([]byte(`<script>self.__next_f.push(</script>`))
	assert.Error(t, err)

	// Binary chunks are base64-decoded.
	payload, err = ExtractNextPayload([]byte(`<script>self.__next_f.push([1,"0:[\"a"])</script>` +
		`<script>self.__next_f.push([3,"YmM="])</script>` +
		`<script>self.__next_f.push([1,"\"]\n"])</script>`))
	require.NoError(t, err)
	assert.Equal(t, "0:[\"abc\"]\n", payload)
}

func TestExtractNextPayloadTimeout(t *testing.T) {
	_, err := ExtractNextPayload([]byte(`<script>self.__next_f; while (true) {}</script>`))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestFindObject(t *testing.T) {
	type player struct {
		Src      string `json:"src"`
		Poster   string `json:"poster"`
		MovieSeq int    `json:"movieSeq"`
		Title    string `json:"title"` // not in the payloads.
	}

	find := func(payload string, opts ...Option) (p player) {
		require.NoError(t, FindObject(payload, &p, opts...))
		return
	}

	assert.Equal(t, player{
		Src:      "https://example.com/a.m3u8",
		Poster:   "https://example.com/a.jpg",
		MovieSeq: 1,
	}, find(samplePayload))

	// Objects without any field of v never match.
	var unknown struct{ Foo string }
	assert.EqualError(t, FindObject(samplePayload, &unknown), "rsc: object not found: *struct { Foo string }")
	assert.ErrorIs(t, FindObject(samplePayload, &unknown, RequireKeys("src")), ErrNotFound)

	// The second object matches more fields.
	const payload = `0:[{"src":"a.jpg","alt":"a"},{"src":"b.m3u8","poster":"b.jpg","movieSeq":2}]` + "\n"
	assert.Equal(t, player{Src: "b.m3u8", Poster: "b.jpg", MovieSeq: 2}, find(payload))

	// Required keys narrow down the candidates.
	assert.Equal(t, player{Src: "a.jpg"}, find(payload, RequireKeys("alt")))
	assert.ErrorIs(t, FindObject(payload, new(player), RequireKeys("missing")), ErrNotFound)

	// Objects tied for best that decode to the same value are one match.
	assert.Equal(t, player{Src: "a", Poster: "b"},
		find(`0:{"src":"a","poster":"b","id":1}`+"\n"+`1:[{"src":"a","poster":"b","id":2}]`+"\n"))
	assert.EqualError(t, FindObject(`0:[{"src":"a","poster":"b"},{"src":"c","poster":"d"}]`+"\n",
		new(player)), "rsc: multiple objects match: *rsc.player")
	assert.Equal(t, player{Src: "c", Poster: "d", MovieSeq: 1},
		find(`0:[{"src":"a","poster":"b"},{"src":"c","poster":"d","movieSeq":1}]`+"\n"))

	// A map has no fields, so keys are required.
	var m map[string]any
	assert.Error(t, FindObject(payload, &m))
	require.NoError(t, FindObject(payload, &m, RequireKeys("poster")))
	assert.Equal(t, "b.m3u8", m["src"])

	// v must be a non-nil pointer.
	assert.Error(t, FindObject(payload, player{}))
	assert.Error(t, FindObject(payload, (*player)(nil)))

	// Length-prefixed rows are skipped, even with newlines.
	const text = "a\n7:{\"src\":\"fake\"}\n"
	assert.Equal(t, player{Src: "real"},
		find(fmt.Sprintf("5:T%x,%s6:{\"src\":\"real\"}\n", len(text), text)))
	assert.Equal(t, player{Src: "real"}, find("5:A4,\x00\x01\x02\x036:{\"src\":\"real\"}\n"))

	// Names are matched and decoded case-sensitively.
	assert.Equal(t, player{Src: "a"},
		find(`0:{"SRC":"x","POSTER":"y"}`+"\n"+`1:{"src":"a","Poster":"b"}`+"\n"))
	assert.ErrorIs(t, FindObject(`0:{"src":"a"}`+"\n", new(player), RequireKeys("SRC")), ErrNotFound)
	var untagged struct{ Src string }
	require.NoError(t, FindObject(`0:{"src":"x"}`+"\n"+`1:{"Src":"a"}`+"\n", &untagged))
	assert.Equal(t, "a", untagged.Src)

	// v is left untouched on failure.
	var tagged struct {
		Src  string   `json:"src"`
		Tags []string `json:"tags"`
	}
	tagged.Src = "untouched"
	var semErr *json.SemanticError
	assert.ErrorAs(t, FindObject(`0:{"src":"a","tags":"$4"}`+"\n", &tagged), &semErr)
	assert.Equal(t, "untouched", tagged.Src)
	assert.Nil(t, tagged.Tags)

	// The object is merged into v.
	merged := player{Title: "default"}
	require.NoError(t, FindObject(`0:{"src":"a","poster":"b"}`+"\n", &merged))
	assert.Equal(t, player{Src: "a", Poster: "b", Title: "default"}, merged)

	kept := map[string]any{"keep": 1}
	require.NoError(t, FindObject(`0:{"src":"a"}`+"\n", &kept, RequireKeys("src")))
	assert.Equal(t, map[string]any{"keep": 1, "src": "a"}, kept)
}

func TestFindObjectDecodeValues(t *testing.T) {
	var v struct {
		Src   string    `json:"src"`
		Price string    `json:"price"`
		Date  time.Time `json:"date"`
		Size  int64     `json:"size"`
		NaN   *float64  `json:"nan"`
	}
	require.NoError(t, FindObject(`0:{"src":"$undefined","price":"$$5",`+
		`"date":"$D2026-11-23T00:00:00.000Z","size":"$n123","nan":"$NaN"}`+"\n", &v))
	assert.Empty(t, v.Src)
	assert.Equal(t, "$5", v.Price)
	assert.Equal(t, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC), v.Date)
	assert.Equal(t, int64(123), v.Size)
	assert.Nil(t, v.NaN)
}
