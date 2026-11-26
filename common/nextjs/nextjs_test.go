package nextjs

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHTML = `<html><head>
<script src="/_next/static/chunks/main.js" async=""></script>
</head><body>
<script>document.write("not a flight data script")</script>
<script>(self.__next_f=self.__next_f||[]).push([0])</script>
<script>self.__next_f.push([1,"1:\"$Sreact.fragment\"\n2:I[123,[],\"default\"]\n"])</script>
<script>self.__next_f.push([1,"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\","])</script>
<script>self.__next_f.push([1,"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n"])</script>
<script>self.__next_f.push([2,"ignored"])</script>
</body></html>`

const testPayload = "1:\"$Sreact.fragment\"\n" +
	"2:I[123,[],\"default\"]\n" +
	"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\"," +
	"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n"

func TestExtractPayload(t *testing.T) {
	payload, err := ExtractPayload([]byte(testHTML))
	require.NoError(t, err)
	assert.Equal(t, testPayload, payload)

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
	if assert.True(t, FindObject(testPayload, &props, "src", "poster")) {
		assert.Equal(t, "https://example.com/a.m3u8", props.Src)
		assert.Equal(t, "https://example.com/a.jpg", props.Poster)
		assert.Equal(t, 1, props.MovieSeq)
	}

	var v map[string]any
	assert.False(t, FindObject(testPayload, &v, "src", "missing"))
	assert.False(t, FindObject(testPayload, &v))
}

func extractTestdata(t *testing.T, name string) string {
	t.Helper()
	page, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	payload, err := ExtractPayload(page)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	return payload
}

// testdata/heyzo.html is the HEYZO movie page https://www.heyzo.com/movie/3986.
func TestHeyzoPage(t *testing.T) {
	payload := extractTestdata(t, "heyzo.html")

	var player struct {
		Src    string `json:"src"`
		Type   string `json:"type"`
		Poster string `json:"poster"`
	}
	if assert.True(t, FindObject(payload, &player, "src", "poster")) {
		assert.Equal(t, "https://hls.heyzo.com/sample/3000/3986/mb.m3u8", player.Src)
		assert.Equal(t, "application/x-mpegURL", player.Type)
		assert.Equal(t, "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg", player.Poster)
	}

	var movie struct {
		MovieSeq int    `json:"movie_seq"`
		MovieID  string `json:"movie_id"`
		Name     string `json:"name"`
	}
	if assert.True(t, FindObject(payload, &movie, "movie_seq", "movie_id")) {
		assert.Equal(t, 4001, movie.MovieSeq)
		assert.Equal(t, "3986", movie.MovieID)
		assert.Equal(t, "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2", movie.Name)
	}
}

// testdata/fanza.html is the FANZA search result page of `SSIS-419`.
func TestFanzaPage(t *testing.T) {
	payload := extractTestdata(t, "fanza.html")

	var resp struct {
		BackendResponse struct {
			SearchWord string `json:"search_word"`
			Contents   struct {
				Count int `json:"count"`
				Data  []struct {
					ContentID string   `json:"content_id"`
					Casts     []string `json:"casts"`
				} `json:"data"`
			} `json:"contents"`
		} `json:"backendResponse"`
	}
	require.True(t, FindObject(payload, &resp, "backendResponse"))

	r := resp.BackendResponse
	assert.Equal(t, "SSIS-419", r.SearchWord)
	assert.Equal(t, 3, r.Contents.Count)
	var ids []string
	for _, data := range r.Contents.Data {
		ids = append(ids, data.ContentID)
	}
	assert.Equal(t, []string{"ssis419", "9ssis419", "4ssis419"}, ids)
	assert.Equal(t, []string{"三上悠亜"}, r.Contents.Data[0].Casts)
}
