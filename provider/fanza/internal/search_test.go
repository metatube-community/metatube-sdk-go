package internal

import (
	"embed"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed sample.html
var _sampleFS embed.FS

func TestParseSearchPage(t *testing.T) {
	p := &http.Transport{}
	p.RegisterProtocol("file", http.NewFileTransport(http.FS(_sampleFS)))

	c := colly.NewCollector()
	c.WithTransport(p)

	parser := NewSearchPageParser()

	c.OnXML("//script", func(e *colly.XMLElement) {
		assert.NoError(t, parser.LoadJSCode(e.Text))
	})

	err := c.Visit("file:///sample.html")
	require.NoError(t, err)

	resp := &ResponseWrapper{}
	err = parser.Parse(resp)
	require.NoError(t, err)
	require.NotZero(t, resp.BackendResponse.Contents.Count)

	data, _ := json.MarshalIndent(resp, "", "\t")
	t.Logf("%s", data)
}
