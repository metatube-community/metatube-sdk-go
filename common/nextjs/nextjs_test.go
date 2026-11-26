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
	var props struct {
		Src      string `json:"src"`
		Poster   string `json:"poster"`
		MovieSeq int    `json:"movieSeq"`
	}
	require.True(t, FindObject(samplePayload, &props, "src", "poster"))
	assert.Equal(t, "https://example.com/a.m3u8", props.Src)
	assert.Equal(t, "https://example.com/a.jpg", props.Poster)
	assert.Equal(t, 1, props.MovieSeq)

	var v map[string]any
	assert.False(t, FindObject(samplePayload, &v, "src", "missing"))
	assert.False(t, FindObject(samplePayload, &v))
}
