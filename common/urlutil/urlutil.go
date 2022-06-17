package urlutil

import (
	pkgurl "net/url"

	"github.com/nlnwa/whatwg-url/url"
)

var urlParser = url.NewParser(url.WithPercentEncodeSinglePercentSign())

// Join joins a URL with a path.
func Join(url, path string) string {
	absURL, err := urlParser.ParseRef(url, path)
	if err != nil {
		return ""
	}
	return absURL.Href(false)
}

func MustParse(rawURL string) *pkgurl.URL {
	u, err := pkgurl.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
