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
	payload, err := ExtractPayload([]byte(samplePage))
	require.NoError(t, err)
	assert.Equal(t, samplePayload, payload)

	payload, err = ExtractPayload([]byte(`<html></html>`))
	require.NoError(t, err)
	assert.Empty(t, payload)

	// Failing scripts are skipped.
	payload, err = ExtractPayload([]byte(`<script>self.__next_f.push(</script>` +
		`<script>self.__next_f.push([1,"0:{}\n"]); document.body</script>` +
		`<script>self.__next_f.push([1,"1:[]\n"])</script>`))
	require.NoError(t, err)
	assert.Equal(t, "0:{}\n1:[]\n", payload)

	// Binary chunks are base64-decoded.
	payload, err = ExtractPayload([]byte(`<script>self.__next_f.push([1,"0:[\"a"])</script>` +
		`<script>self.__next_f.push([3,"YmM="])</script>` +
		`<script>self.__next_f.push([1,"\"]\n"])</script>`))
	require.NoError(t, err)
	assert.Equal(t, "0:[\"abc\"]\n", payload)

	// A script error is reported if no data could be extracted at all.
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

	find := func(payload string, keys ...string) (p player) {
		require.True(t, FindObject(payload, &p, keys...))
		return
	}

	assert.Equal(t, player{
		Src:      "https://example.com/a.m3u8",
		Poster:   "https://example.com/a.jpg",
		MovieSeq: 1,
	}, find(samplePayload))

	// Objects without any field of the struct never match, even with keys.
	var unknown struct{ Foo string }
	assert.False(t, FindObject(samplePayload, &unknown))
	assert.False(t, FindObject(samplePayload, &unknown, "src"))

	// The second object matches more fields than the first one.
	const payload = `0:[{"src":"a.jpg","alt":"a"},{"src":"b.m3u8","poster":"b.jpg","movieseq":2}]` + "\n"
	assert.Equal(t, player{Src: "b.m3u8", Poster: "b.jpg", MovieSeq: 2}, find(payload))

	// Keys narrow down the candidates.
	assert.Equal(t, player{Src: "a.jpg"}, find(payload, "alt"))
	assert.False(t, FindObject(payload, new(player), "missing"))

	// A map has no fields to match, so keys are required.
	var m map[string]any
	assert.False(t, FindObject(payload, &m))
	require.True(t, FindObject(payload, &m, "poster"))
	assert.Equal(t, "b.m3u8", m["src"])

	// Keys are compared case-insensitively.
	assert.Equal(t, player{Src: "a.jpg"}, find(payload, "ALT"))

	// v must be a non-nil pointer, which may point to a struct pointer.
	assert.False(t, FindObject(payload, player{}))
	assert.False(t, FindObject(payload, (*player)(nil)))
	var pp *player
	require.True(t, FindObject(payload, &pp))
	assert.Equal(t, &player{Src: "b.m3u8", Poster: "b.jpg", MovieSeq: 2}, pp)

	// Fields of embedded structs are matched too.
	type media struct {
		Src string `json:"src"`
	}
	var embedded struct {
		media
		Poster string `json:"poster"`
	}
	require.True(t, FindObject(payload, &embedded))
	assert.Equal(t, "b.m3u8", embedded.Src)
	assert.Equal(t, "b.jpg", embedded.Poster)

	// Keys with the same name in different cases count once.
	assert.Equal(t, player{Src: "b", Poster: "b"},
		find(`0:[{"src":"a","SRC":"a"},{"src":"b","poster":"b"}]`+"\n"))

	// Objects that cannot be unmarshaled are skipped, and v is left untouched
	// if none can.
	assert.Equal(t, player{Src: "b"},
		find(`0:[{"src":"a","poster":"a","movieseq":"bad"},{"src":"b"}]`+"\n"))
	p := player{Title: "untouched"}
	assert.False(t, FindObject(`0:{"src":"a","movieseq":"bad"}`+"\n", &p))
	assert.Equal(t, player{Title: "untouched"}, p)

	// Length-prefixed rows are skipped by their length, even if they contain
	// newlines, and a truncated one ends the payload.
	const text = "a\n7:{\"src\":\"fake\"}\n"
	assert.Equal(t, player{Src: "real"},
		find(fmt.Sprintf("5:T%x,%s6:{\"src\":\"real\"}\n", len(text), text)))
	assert.Equal(t, player{Src: "real"}, find("5:A4,\x00\x01\x02\x036:{\"src\":\"real\"}\n"))
	assert.False(t, FindObject("5:Tff,short\n6:{\"src\":\"real\"}\n", new(player)))
}

func TestFindObjectDecodeValues(t *testing.T) {
	var v struct {
		Src      string    `json:"src"`
		Price    string    `json:"price"`
		Date     time.Time `json:"date"`
		Size     int64     `json:"size"`
		Zero     float64   `json:"zero"`
		NaN      *float64  `json:"nan"`
		Infinity *float64  `json:"infinity"`
	}
	require.True(t, FindObject(`0:{"src":"$undefined","price":"$$5","date":"$D2026-11-23T00:00:00.000Z",`+
		`"size":"$n123","zero":"$-0","nan":"$NaN","infinity":"$Infinity"}`+"\n", &v))
	assert.Empty(t, v.Src)
	assert.Equal(t, "$5", v.Price)
	assert.Equal(t, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC), v.Date)
	assert.Equal(t, int64(123), v.Size)
	assert.Zero(t, v.Zero)
	assert.Nil(t, v.NaN)
	assert.Nil(t, v.Infinity)

	// Bigints are decoded into strings if needed.
	var id struct {
		ID string `json:"id"`
	}
	require.True(t, FindObject(`0:{"id":"$n12345678901234567890"}`+"\n", &id))
	assert.Equal(t, "12345678901234567890", id.ID)
}

func TestFindObjectReferences(t *testing.T) {
	type movie struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
		Actor   struct {
			Name string `json:"name"`
		} `json:"actor"`
		Tags []string `json:"tags"`
	}
	const summary = "a long summary\nsent as a text row"
	payload := fmt.Sprintf("1f:T%x,%s", len(summary), summary) +
		`3:["tag-a","tag-b"]` + "\n" +
		`6:["$","main",null,{"children":[["$","$L4",null,{"actor":{"name":"x"}}]]}]` + "\n" +
		`7:{"title":"t","summary":"$1f","actor":"$6:props:children:0:props:actor","tags":"$@3"}` + "\n"

	var m movie
	require.True(t, FindObject(payload, &m))
	assert.Equal(t, "t", m.Title)
	assert.Equal(t, summary, m.Summary)
	assert.Equal(t, "x", m.Actor.Name)
	assert.Equal(t, []string{"tag-a", "tag-b"}, m.Tags)

	// Unresolvable and cyclic references are kept as they are.
	var v struct {
		Missing string `json:"missing"`
		Cyclic  string `json:"cyclic"`
	}
	require.True(t, FindObject(`8:"$8"`+"\n"+`9:{"missing":"$ff","cyclic":"$8"}`+"\n", &v))
	assert.Equal(t, "$ff", v.Missing)
	assert.Equal(t, "$8", v.Cyclic)
}
