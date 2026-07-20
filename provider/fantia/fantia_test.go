package fantia

import "testing"

func TestSharedSessionID(t *testing.T) {
	t.Setenv("FANTIA_SESSION_ID", "fantia_session")
	if NewPost().sessionID != "fantia_session" || NewProduct().sessionID != "fantia_session" {
		t.Fatal("shared Fantia session ID was not applied")
	}
}

func TestNormalizeAndParseMovieID(t *testing.T) {
	tests := []struct {
		provider *Fantia
		id       string
		rawURL   string
	}{
		{NewPost(), "FANTIA-POST-123", "https://fantia.jp/posts/123"},
		{NewPost(), "fantia-posts-3906063", "https://fantia.jp/posts/3906063"},
		{NewProduct(), "FANTIA-PRODUCT-2132423", "https://fantia.jp/products/2132423?locale=jp"},
	}
	for _, test := range tests {
		id := test.provider.NormalizeMovieID(test.id)
		parsed, err := test.provider.ParseMovieIDFromURL(test.rawURL)
		if err != nil || parsed != id {
			t.Fatalf("normalize=%q parse=%q error=%v", id, parsed, err)
		}
	}
	if NewPost().NormalizeMovieID("FANTIA-PRODUCT-123") != "" {
		t.Fatal("post provider accepted a product ID")
	}
	if _, err := NewPost().ParseMovieIDFromURL("https://example.com/posts/123"); err == nil {
		t.Fatal("accepted a non-Fantia URL")
	}
}

func TestProductNumber(t *testing.T) {
	info := NewProduct().newMovieInfo("2132423")
	if info.Number != "FANTIA-PRODUCT-2132423" {
		t.Fatalf("unexpected number: %q", info.Number)
	}
}

func TestDecodeProducts(t *testing.T) {
	products := decodeProducts(`[{"@type":"Product","name":"Title","description":"Summary","image":["/cover.jpg","/sample.jpg"],"brand":{"name":"Creator"}}]`)
	if len(products) != 1 || products[0].Name != "Title" || products[0].Brand.Name != "Creator" || len(products[0].Image) != 2 {
		t.Fatalf("unexpected products: %#v", products)
	}
}

func TestPostMovieInfo(t *testing.T) {
	var post postData
	post.ID = 123
	post.Title = "Title"
	post.PostedAt = "Mon, 20 Jul 2026 12:00:00 +0900"
	post.Thumb.Original = "/cover.jpg"
	post.Fanclub.Name = "Fanclub"
	post.Fanclub.User.Name = "Creator"
	post.Rating = "adult"
	post.Tags = append(post.Tags, postTag{Name: "Tag"})
	post.PostContents = append(post.PostContents, postContent{
		Category:      "blog",
		Comment:       `{"ops":[{"insert":"Blog text"},{"insert":{"fantiaImage":{"original_url":"/blog.jpg"}}}]}`,
		DownloadURI:   "/movie.mp4",
		VisibleStatus: "visible",
	})

	info := NewPost().postMovieInfo("123", &post)
	if info.Title != "Title" || info.Summary != "Blog text" || info.CoverURL != rootURL+"/cover.jpg" || info.Maker != "Fanclub" || info.Label != "adult" {
		t.Fatalf("unexpected info: %#v", info)
	}
	if info.PreviewVideoURL != rootURL+"/movie.mp4" {
		t.Fatalf("unexpected video: %q", info.PreviewVideoURL)
	}
	if len(info.Actors) != 1 || info.Actors[0] != "Creator" || len(info.Genres) != 1 || len(info.PreviewImages) != 1 {
		t.Fatalf("unexpected lists: actors=%v genres=%v images=%v", info.Actors, info.Genres, info.PreviewImages)
	}
}
