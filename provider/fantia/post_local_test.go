package fantia

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

func testPost(t *testing.T, respond func(*http.Request) (string, string)) *Post {
	t.Helper()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, contentType := respond(req)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{contentType}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	return &Post{
		Scraper:   scraper.NewDefaultScraper(PostName, postBaseURL, Priority, language.Japanese, scraper.WithTransport(transport)),
		sessionID: "test-session",
	}
}

func TestFantiaPostFetchesCompleteVisibleContent(t *testing.T) {
	requests := 0
	p := testPost(t, func(req *http.Request) (string, string) {
		requests++
		if cookie, err := req.Cookie("_session_id"); err != nil || cookie.Value != "test-session" {
			t.Errorf("Fantia session cookie missing on %s", req.URL)
		}
		switch req.URL.String() {
		case baseURL:
			return `<html><head><meta name="csrf-token" content="csrf-test"></head></html>`, "text/html; charset=utf-8"
		case "https://fantia.jp/api/v1/posts/123":
			if req.Header.Get("X-CSRF-Token") != "csrf-test" || req.Header.Get("X-Requested-With") != "XMLHttpRequest" {
				t.Errorf("missing Fantia API headers: %v", req.Header)
			}
			return `{"post":{"id":123,"title":"Test post","comment":"Short introduction","rating":"adult","posted_at":"Mon, 20 Jul 2026 12:00:00 +0900","thumb":{"original":"/cover.jpg"},"fanclub":{"name":"Fanclub","user":{"name":"Creator"}},"tags":[{"name":"Tag"},{"name":"Tag"}],"post_contents":[{"category":"blog","visible_status":"visible","comment":"{\"ops\":[{\"insert\":\"First line\\nSecond line\"},{\"insert\":{\"fantiaImage\":{\"original_url\":\"/blog.jpg\"}}}]}","download_uri":"/sample.mp4","post_content_photos":[{"url":{"original":"/photo.jpg"}}]},{"category":"blog","visible_status":"hidden","comment":"{\"ops\":[{\"insert\":\"Hidden text\"}]}"}]}}`, "application/json"
		default:
			t.Errorf("unexpected request: %s", req.URL)
			return "", "text/plain"
		}
	})

	info, err := p.GetMovieInfoByURL("https://fantia.jp/posts/123?tracking=ignored")
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Errorf("got %d requests, want 2", requests)
	}
	if info.Provider != PostName || info.Number != "FANTIA-POST-123" || info.Homepage != "https://fantia.jp/posts/123" {
		t.Errorf("unexpected post identity: %+v", info)
	}
	if info.Summary != "Short introduction\n\nFirst line\nSecond line" {
		t.Errorf("unexpected full summary: %q", info.Summary)
	}
	if info.CoverURL != "https://fantia.jp/cover.jpg" || info.PreviewVideoURL != "https://fantia.jp/sample.mp4" {
		t.Errorf("unexpected post media: %+v", info)
	}
	if len(info.PreviewImages) != 2 || info.PreviewImages[0] != "https://fantia.jp/photo.jpg" || info.PreviewImages[1] != "https://fantia.jp/blog.jpg" {
		t.Errorf("unexpected post images: %v", info.PreviewImages)
	}
	if len(info.Genres) != 1 || info.Genres[0] != "Tag" || len(info.Actors) != 1 || info.Actors[0] != "Creator" {
		t.Errorf("unexpected post metadata: %+v", info)
	}
}

func TestFantiaPostRejectsForeignURLBeforeSendingCookie(t *testing.T) {
	called := false
	p := testPost(t, func(*http.Request) (string, string) {
		called = true
		return "", "text/plain"
	})
	for _, rawURL := range []string{
		"https://example.com/posts/123",
		"http://fantia.jp/posts/123",
		"https://fantia.jp.evil.example/posts/123",
		"https://fantia.jp:443/posts/123",
		"https://fantia.jp/products/123",
		"https://fantia.jp/posts/not-an-id",
	} {
		if _, err := p.GetMovieInfoByURL(rawURL); err == nil {
			t.Errorf("accepted invalid post URL: %s", rawURL)
		}
	}
	if called {
		t.Fatal("sent a request for an invalid post URL")
	}
}

func TestFantiaSharedSessionFromEnvironment(t *testing.T) {
	t.Setenv("FANTIA_SESSION_ID", "  from-env  ")
	if New().sessionID != "from-env" || NewPost().sessionID != "from-env" {
		t.Fatal("product and post providers did not share the environment session")
	}
}
