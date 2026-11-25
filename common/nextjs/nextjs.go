// Package nextjs provides helpers to extract data from pages rendered by
// the Next.js App Router, which streams the React Server Components (RSC)
// payload to the browser via `self.__next_f.push([1, "..."])` scripts.
package nextjs

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"golang.org/x/net/html"
)

const flightDataVar = "self.__next_f"

// scriptTimeout limits the total execution time of the page scripts.
var scriptTimeout = 3 * time.Second

var errScriptTimeout = errors.New("nextjs: script execution timed out")

// ExtractPayload extracts the RSC payload from a Next.js App Router HTML
// page. It runs the inline flight data scripts of the page in a JS VM with
// a predefined `self.__next_f` array, then concatenates the pushed payload
// chunks.
func ExtractPayload(page []byte) (payload string, err error) {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return "", err
	}

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)

	timer := time.AfterFunc(scriptTimeout, func() {
		vm.Interrupt <- func() { panic(errScriptTimeout) }
	})
	defer timer.Stop()

	defer func() {
		if r := recover(); r != nil {
			if r != errScriptTimeout {
				panic(r)
			}
			err = errScriptTimeout
		}
	}()

	if _, err = vm.Run(`var self = { __next_f: [] };`); err != nil {
		return "", err
	}
	for _, script := range inlineScripts(doc) {
		// Only run the flight data scripts, other scripts may depend
		// on browser APIs that are not available in the VM.
		if !strings.Contains(script, flightDataVar) {
			continue
		}
		if _, err = vm.Run(script); err != nil {
			return "", err
		}
	}

	// Only type 1 chunks, i.e., `[1, "..."]`, carry the payload data.
	v, err := vm.Run(`self.__next_f.filter(function (c) {
		return c && c[0] === 1 && typeof c[1] === "string";
	}).map(function (c) {
		return c[1];
	}).join("");`)
	if err != nil {
		return "", err
	}
	return v.ToString()
}

// inlineScripts returns the contents of all inline <script> elements.
func inlineScripts(doc *html.Node) (scripts []string) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" && !hasAttr(n, "src") {
			var b strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					b.WriteString(c.Data)
				}
			}
			if b.Len() > 0 {
				scripts = append(scripts, b.String())
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return
}

func hasAttr(n *html.Node, key string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// FindObject walks through every JSON row of the RSC payload and unmarshals
// the first found object that contains all the given keys into v. It returns
// false if no such object is found.
func FindObject(payload string, v any, keys ...string) bool {
	for _, line := range strings.Split(payload, "\n") {
		data, ok := parseRow(line)
		if !ok {
			continue
		}
		var row any
		if json.Unmarshal([]byte(data), &row) != nil {
			continue
		}
		if obj := findObject(row, keys); obj != nil {
			b, err := json.Marshal(obj)
			if err != nil {
				return false
			}
			return json.Unmarshal(b, v) == nil
		}
	}
	return false
}

// parseRow parses a payload row in `<hex id>:<data>` format and returns its
// data if it is a JSON array or object, e.g., `6:["$","main",null,{...}]`.
func parseRow(line string) (string, bool) {
	id, data, ok := strings.Cut(line, ":")
	if !ok || id == "" || data == "" {
		return "", false
	}
	if _, err := strconv.ParseUint(id, 16, 64); err != nil {
		return "", false
	}
	if data[0] != '[' && data[0] != '{' {
		return "", false
	}
	return data, true
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
