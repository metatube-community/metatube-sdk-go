package nextjs

import (
	"embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Inline samples
// ---------------------------------------------------------------------------

// samplePage is a minimal Next.js App Router page. It covers:
//   - an external script and a non-flight inline script, which must not run;
//   - the `self.__next_f` bootstrap script;
//   - a payload row split across two chunks;
//   - a non-type-1 chunk, which must be ignored.
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

// samplePayload is the payload extracted from samplePage.
const samplePayload = "1:\"$Sreact.fragment\"\n" +
	"2:I[123,[],\"default\"]\n" +
	"6:[\"$\",\"main\",null,{\"children\":[[\"$\",\"$L2\",null,{\"src\":\"https://example.com/a.m3u8\"," +
	"\"poster\":\"https://example.com/a.jpg\",\"movieSeq\":1}]]}]\n"

type sampleProps struct {
	Src      string `json:"src"`
	Poster   string `json:"poster"`
	MovieSeq int    `json:"movieSeq"`
}

// ---------------------------------------------------------------------------
// Unit tests
// ---------------------------------------------------------------------------

func TestExtractPayload(t *testing.T) {
	for _, unit := range []struct {
		name    string
		page    string
		want    string
		wantErr bool
	}{
		{name: "sample page", page: samplePage, want: samplePayload},
		{name: "no flight data", page: `<html></html>`, want: ""},
		{name: "script syntax error", page: `<script>self.__next_f.push(</script>`, wantErr: true},
	} {
		t.Run(unit.name, func(t *testing.T) {
			payload, err := ExtractPayload([]byte(unit.page))
			if unit.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, unit.want, payload)
		})
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
		name  string
		keys  []string
		found bool
		want  sampleProps
	}{
		{
			name:  "found",
			keys:  []string{"src", "poster"},
			found: true,
			want: sampleProps{
				Src:      "https://example.com/a.m3u8",
				Poster:   "https://example.com/a.jpg",
				MovieSeq: 1,
			},
		},
		{name: "missing key", keys: []string{"src", "missing"}},
		{name: "no keys"},
	} {
		t.Run(unit.name, func(t *testing.T) {
			var got sampleProps
			require.Equal(t, unit.found, FindObject(samplePayload, &got, unit.keys...))
			assert.Equal(t, unit.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Saved pages
// ---------------------------------------------------------------------------

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

const (
	heyzoMoviePage  = "55307bfd703df11319fecbfed69883f5" // https://www.heyzo.com/movie/3986
	fanzaSearchPage = "a0ef9a6847f4d22cca6abc9480b7a3ae" // https://www.dmm.co.jp/search/=/searchstr=SSIS-419/limit=30/sort=rankprofile
)

type heyzoPlayer struct {
	Src    string `json:"src"`
	Type   string `json:"type"`
	Poster string `json:"poster"`
}

type heyzoMovie struct {
	MovieSeq int    `json:"movie_seq"`
	MovieID  string `json:"movie_id"`
	Name     string `json:"name"`
}

type fanzaSearch struct {
	BackendResponse fanzaBackendResponse `json:"backendResponse"`
}

type fanzaBackendResponse struct {
	SearchWord string        `json:"search_word"`
	Contents   fanzaContents `json:"contents"`
}

type fanzaContents struct {
	Count int            `json:"count"`
	Data  []fanzaProduct `json:"data"`
}

type fanzaProduct struct {
	ContentID string   `json:"content_id"`
	Casts     []string `json:"casts"`
}

func TestSavedPages(t *testing.T) {
	for _, unit := range []struct {
		name string
		page string
		keys []string
		got  any // pointer to the zero value to unmarshal into.
		want any
	}{
		{
			name: "HEYZO movie player",
			page: heyzoMoviePage,
			keys: []string{"src", "poster"},
			got:  &heyzoPlayer{},
			want: &heyzoPlayer{
				Src:    "https://hls.heyzo.com/sample/3000/3986/mb.m3u8",
				Type:   "application/x-mpegURL",
				Poster: "https://www.heyzo.com/contents/3000/3986/images/player_thumbnail.jpg",
			},
		},
		{
			name: "HEYZO movie info",
			page: heyzoMoviePage,
			keys: []string{"movie_seq", "movie_id"},
			got:  &heyzoMovie{},
			want: &heyzoMovie{
				MovieSeq: 4001,
				MovieID:  "3986",
				Name:     "男優のテクをガマンできたらご褒美ゲット！でも、ダメだったら。。。Vol.2",
			},
		},
		{
			name: "FANZA search result",
			page: fanzaSearchPage,
			keys: []string{"backendResponse"},
			got:  &fanzaSearch{},
			want: &fanzaSearch{
				BackendResponse: fanzaBackendResponse{
					SearchWord: "SSIS-419",
					Contents: fanzaContents{
						Count: 3,
						Data: []fanzaProduct{
							{ContentID: "ssis419", Casts: []string{"三上悠亜"}},
							{ContentID: "9ssis419", Casts: []string{"三上悠亜"}},
							{ContentID: "4ssis419", Casts: []string{"三上悠亜"}},
						},
					},
				},
			},
		},
	} {
		t.Run(unit.name, func(t *testing.T) {
			payload := extractSavedPage(t, unit.page)
			require.True(t, FindObject(payload, unit.got, unit.keys...))
			assert.Equal(t, unit.want, unit.got)
		})
	}
}

// extractSavedPage decodes a base64-encoded page from testdata and returns
// its RSC payload.
func extractSavedPage(t *testing.T, filename string) string {
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
