package nextjs

import (
	"embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
  # Generate MD5-hashed pages:
  for PAGE in *.html; do
    gbase64 "$PAGE" > "$(md5sum "$PAGE" | cut -d' ' -f1)"
  done

  # Decode page:
  gbase64 -d "$PAGE" > "$PAGE".html
*/

//go:embed testdata/*
var testdata embed.FS

// Saved Next.js App Router pages in testdata.
const (
	heyzoMoviePage  = "55307bfd703df11319fecbfed69883f5" // https://www.heyzo.com/movie/3986
	fanzaSearchPage = "a0ef9a6847f4d22cca6abc9480b7a3ae" // https://www.dmm.co.jp/search/=/searchstr=SSIS-419/limit=30/sort=rankprofile
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

func extractTestdata(t *testing.T, filename string) string {
	t.Helper()
	data, err := testdata.ReadFile("testdata/" + filename)
	require.NoError(t, err)
	page, err := base64.StdEncoding.DecodeString(string(data))
	require.NoError(t, err)
	payload, err := ExtractPayload(page)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	return payload
}

func TestHeyzoMoviePage(t *testing.T) {
	payload := extractTestdata(t, heyzoMoviePage)

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

func TestFanzaSearchPage(t *testing.T) {
	payload := extractTestdata(t, fanzaSearchPage)

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
