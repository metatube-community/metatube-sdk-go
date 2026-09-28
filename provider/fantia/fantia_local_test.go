package fantia

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func testFantia(t *testing.T, html string, check func(*http.Request)) *Fantia {
	t.Helper()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		check(req)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			Body:       io.NopCloser(strings.NewReader(html)),
			Request:    req,
		}, nil
	})
	return &Fantia{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese,
			scraper.WithTransport(transport)),
		sessionID: "test-session",
	}
}

func TestFantiaRejectsForeignURLBeforeSendingCookie(t *testing.T) {
	called := false
	ft := testFantia(t, "", func(*http.Request) { called = true })
	for _, rawURL := range []string{
		"https://example.com/products/123",
		"http://fantia.jp/products/123",
		"https://fantia.jp.evil.example/products/123",
		"https://fantia.jp:443/products/123",
		"https://fantia.jp/posts/123",
		"https://fantia.jp/products/not-an-id",
	} {
		if _, err := ft.GetMovieInfoByURL(rawURL); err == nil {
			t.Errorf("GetMovieInfoByURL(%q) should reject invalid URL", rawURL)
		}
	}
	if called {
		t.Fatal("sent a request for an invalid URL")
	}
}

func TestFantiaDescriptionPrefersFullTextAndPreservesBreaks(t *testing.T) {
	page := `<html><head><script type="application/ld+json">[{"@type":"Product","description":"Short..."}]</script></head><body>
<h1 class="product-title">Test product</h1>
<div class="product-description"><h3 class="content-title">About product</h3><div><p>First line<br>Second line</p><p>Final line</p></div></div>
</body></html>`
	ft := testFantia(t, page, func(req *http.Request) {
		if req.URL.String() != "https://fantia.jp/products/123" {
			t.Errorf("unexpected request URL: %s", req.URL)
		}
		if cookie, err := req.Cookie("_session_id"); err != nil || cookie.Value != "test-session" {
			t.Error("Fantia session cookie was not sent to Fantia")
		}
	})
	info, err := ft.GetMovieInfoByURL("https://fantia.jp/products/123?tracking=ignored")
	if err != nil {
		t.Fatal(err)
	}
	if info.Summary != "First line\nSecond line\nFinal line" {
		t.Errorf("unexpected summary: %q", info.Summary)
	}
	if info.Homepage != "https://fantia.jp/products/123" {
		t.Errorf("unexpected homepage: %q", info.Homepage)
	}
}

func TestFantiaDescriptionUsesLongerJSONLD(t *testing.T) {
	page := `<html><head><script type="application/ld+json">[{"@type":"Product","description":"Complete JSON description"}]</script></head><body>
<div class="product-description"><h3 class="content-title">About product</h3><p>Short...</p></div>
</body></html>`
	ft := testFantia(t, page, func(*http.Request) {})
	info, err := ft.GetMovieInfoByID("FANTIA-123")
	if err != nil {
		t.Fatal(err)
	}
	if info.Summary != "Complete JSON description" {
		t.Errorf("unexpected summary: %q", info.Summary)
	}
}

func TestFantiaProductSchemaObjectKeepsFullDescription(t *testing.T) {
	page := `<html><head><meta property="og:description" content="Preview..."><script type="application/ld+json">{"@type":"Product","name":"Schema title","description":"Short schema description","category":"Illustration","brand":{"name":"Creator"},"image":"/cover.jpg"}</script></head><body>
<h1 class="product-title">Product title</h1><div class="product-description"><p>The complete product description is longer than the preview and the schema description.</p></div>
</body></html>`
	ft := testFantia(t, page, func(*http.Request) {})
	info, err := ft.GetMovieInfoByID("123")
	if err != nil {
		t.Fatal(err)
	}
	if info.Summary != "The complete product description is longer than the preview and the schema description." {
		t.Errorf("unexpected product summary: %q", info.Summary)
	}
	if info.Maker != "Creator" || info.CoverURL != "https://fantia.jp/cover.jpg" || len(info.Genres) != 1 || info.Genres[0] != "Illustration" {
		t.Errorf("unexpected schema metadata: %+v", info)
	}
}
