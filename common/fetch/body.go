package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

func WithJSONBody(v any) io.Reader {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		panic(err)
	}
	return buf
}

func WithURLEncodedBody(query map[string]string) io.Reader {
	v := &url.Values{}
	for key, value := range query {
		v.Set(key, value)
	}
	return strings.NewReader(v.Encode())
}
