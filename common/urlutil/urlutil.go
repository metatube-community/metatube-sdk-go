package urlutil

import (
	"github.com/nlnwa/whatwg-url/url"
)

var urlParser = url.NewParser(url.WithPercentEncodeSinglePercentSign())

// JoinURL joins a URL with a path.
func JoinURL(url, path string) string {
	absURL, err := urlParser.ParseRef(url, path)
	if err != nil {
		return ""
	}
	return absURL.Href(false)
}
