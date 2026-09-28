package engine

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider"
)

func TestFantiaMovieProviderRegistered(t *testing.T) {
	found := map[string]bool{}
	provider.RangeMovieFactory(func(name string, factory provider.MovieFactory) bool {
		if name == "Fantia" || name == "FantiaPost" {
			found[name] = true
			if got := factory().URL().Hostname(); got != "fantia.jp" {
				t.Errorf("%s provider host = %q", name, got)
			}
		}
		return true
	})
	for _, name := range []string{"Fantia", "FantiaPost"} {
		if !found[name] {
			t.Errorf("%s movie provider is not registered in the engine", name)
		}
	}
}

func TestFantiaURLRouting(t *testing.T) {
	e := New(nil)
	for _, test := range []struct{ rawURL, want string }{
		{"https://fantia.jp/products/123", "Fantia"},
		{"https://fantia.jp/posts/123", "FantiaPost"},
	} {
		p, err := e.GetMovieProviderByURL(test.rawURL)
		if err != nil {
			t.Errorf("route %s: %v", test.rawURL, err)
		} else if p.Name() != test.want {
			t.Errorf("route %s to %s, want %s", test.rawURL, p.Name(), test.want)
		}
	}
}
