// Package nextjs provides helpers to extract data from pages rendered by
// the Next.js App Router, which streams the React Server Components (RSC)
// payload to the browser via `self.__next_f.push([1, "..."])` scripts.
package nextjs

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	flightDataRe = regexp.MustCompile(`self\.__next_f\.push\((\[.*?\])\)</script>`)
	payloadRowRe = regexp.MustCompile(`^[0-9a-f]+:([\[{].*)$`)
)

// ExtractPayload extracts and concatenates the RSC payload chunks from a
// Next.js App Router HTML page.
func ExtractPayload(html []byte) string {
	var b strings.Builder
	for _, sub := range flightDataRe.FindAllSubmatch(html, -1) {
		var chunk []any
		if json.Unmarshal(sub[1], &chunk) != nil || len(chunk) != 2 {
			continue
		}
		// Only type 1 chunks carry the payload data.
		if t, ok := chunk[0].(float64); !ok || t != 1 {
			continue
		}
		if s, ok := chunk[1].(string); ok {
			b.WriteString(s)
		}
	}
	return b.String()
}

// FindObject walks through every JSON row of the RSC payload and unmarshals
// the first found object that contains all the given keys into v. It returns
// false if no such object is found.
func FindObject(payload string, v any, keys ...string) bool {
	for _, line := range strings.Split(payload, "\n") {
		sub := payloadRowRe.FindStringSubmatch(line)
		if len(sub) != 2 {
			continue
		}
		var row any
		if json.Unmarshal([]byte(sub[1]), &row) != nil {
			continue
		}
		if obj := findObject(row, keys); obj != nil {
			data, err := json.Marshal(obj)
			if err != nil {
				return false
			}
			return json.Unmarshal(data, v) == nil
		}
	}
	return false
}

func findObject(node any, keys []string) map[string]any {
	switch n := node.(type) {
	case map[string]any:
		if hasKeys(n, keys) {
			return n
		}
		for _, child := range n {
			if obj := findObject(child, keys); obj != nil {
				return obj
			}
		}
	case []any:
		for _, child := range n {
			if obj := findObject(child, keys); obj != nil {
				return obj
			}
		}
	}
	return nil
}

func hasKeys(obj map[string]any, keys []string) bool {
	for _, key := range keys {
		if _, ok := obj[key]; !ok {
			return false
		}
	}
	return len(keys) > 0
}
