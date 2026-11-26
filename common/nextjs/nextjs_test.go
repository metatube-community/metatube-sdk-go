package nextjs

import (
	"embed"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pages are base64-encoded and named by their MD5:
//
//	gbase64 page.html > "$(md5sum page.html | cut -d' ' -f1)"
//
//go:embed testdata/*
var testdata embed.FS

const (
	heyzoMoviePage  = "55307bfd703df11319fecbfed69883f5" // https://www.heyzo.com/movie/3986
	fanzaSearchPage = "a0ef9a6847f4d22cca6abc9480b7a3ae" // https://www.dmm.co.jp/search/=/searchstr=SSIS-419/limit=30/sort=rankprofile
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
	heyzo := loadPayload(t, heyzoMoviePage)
	fanza := loadPayload(t, fanzaSearchPage)

	for _, unit := range []struct {
		payload string
		keys    []string
		want    map[string]string // fields of the found object; nil if not found.
	}{
		{samplePayload, []string{"src", "poster"}, map[string]string{
			"src":      "https://example.com/a.m3u8",
			"poster":   "https://example.com/a.jpg",
			"movieSeq": "1",
		}},
		{samplePayload, []string{"src", "missing"}, nil},
		{samplePayload, nil, nil},
		{heyzo, []string{"src", "poster"}, map[string]string{
			"src":    "https://hls.heyzo.com/sample/3000/3986/mb.m3u8",
			"type":   "application/x-mpegURL",
			"poster": "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg",
		}},
		{heyzo, []string{"movie_seq", "movie_id"}, map[string]string{
			"movie_seq": "4001",
			"movie_id":  "3986",
			"name":      "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2",
		}},
		{fanza, []string{"search_word", "contents"}, map[string]string{
			"search_word": "SSIS-419",
			"order_by":    "rankprofile",
			"limit":       "30",
			"page":        "1",
		}},
	} {
		var got map[string]any
		if !assert.Equal(t, unit.want != nil, FindObject(unit.payload, &got, unit.keys...), unit.keys) {
			continue
		}
		for key, want := range unit.want {
			assert.Equal(t, want, fmt.Sprint(got[key]), key)
		}
	}
}

func loadPayload(t *testing.T, filename string) string {
	t.Helper()
	data, err := testdata.ReadFile("testdata/" + filename)
	require.NoError(t, err)
	page, err := base64.StdEncoding.DecodeString(string(data))
	require.NoError(t, err)
	payload, err := ExtractPayload(page)
	require.NoError(t, err)
	return payload
}
