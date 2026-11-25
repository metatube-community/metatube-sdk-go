package nextjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testHTML = `<html><body>
<script>self.__next_f.push([0])</script>
<script>self.__next_f.push([1,"1:\"$Sreact.fragment\"\n2:I[123,[],\"default\"]\n"])</script>
<script>self.__next_f.push([1,"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\","])</script>
<script>self.__next_f.push([1,"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n"])</script>
<script>self.__next_f.push([2,"ignored"])</script>
</body></html>`

func TestExtractPayload(t *testing.T) {
	assert.Equal(t,
		"1:\"$Sreact.fragment\"\n"+
			"2:I[123,[],\"default\"]\n"+
			"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\","+
			"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n",
		ExtractPayload([]byte(testHTML)))
	assert.Empty(t, ExtractPayload([]byte(`<html></html>`)))
}

func TestFindObject(t *testing.T) {
	payload := ExtractPayload([]byte(testHTML))

	var props struct {
		Src      string `json:"src"`
		Poster   string `json:"poster"`
		MovieSeq int    `json:"movieSeq"`
	}
	if assert.True(t, FindObject(payload, &props, "src", "poster")) {
		assert.Equal(t, "https://example.com/a.m3u8", props.Src)
		assert.Equal(t, "https://example.com/a.jpg", props.Poster)
		assert.Equal(t, 1, props.MovieSeq)
	}

	var v map[string]any
	assert.False(t, FindObject(payload, &v, "src", "missing"))
	assert.False(t, FindObject(payload, &v))
}
