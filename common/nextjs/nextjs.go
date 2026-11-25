// Package nextjs provides helpers to extract data from pages rendered by
// the Next.js App Router, which streams the React Server Components (RSC)
// payload to the browser via `self.__next_f.push([1, "..."])` scripts.
package nextjs

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/robertkrimen/otto"
)

// scriptTimeout limits the total execution time of the page scripts.
var scriptTimeout = 3 * time.Second

var errScriptTimeout = errors.New("nextjs: script execution timed out")

// ExtractPayload extracts the RSC payload from a Next.js App Router HTML
// page. It runs the inline flight data scripts of the page in a JS VM with
// a predefined `self.__next_f` array, then concatenates the pushed payload
// chunks.
func ExtractPayload(page []byte) (payload string, err error) {
	doc, err := htmlquery.Parse(bytes.NewReader(page))
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
	// Only run the inline flight data scripts, other scripts may depend
	// on browser APIs that are not available in the VM.
	for _, script := range htmlquery.Find(doc, `//script[not(@src)][contains(., "self.__next_f")]`) {
		if _, err = vm.Run(htmlquery.InnerText(script)); err != nil {
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

// FindObject walks through every JSON row of the RSC payload and unmarshals
// the first found object that contains all the given keys into v. Objects
// are visited depth-first, with map values in sorted key order, so the result
// is deterministic. It returns false if no such object is found.
func FindObject(payload string, v any, keys ...string) bool {
	if len(keys) == 0 {
		return false
	}
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
			// Marshaling a value decoded from JSON never fails.
			b, _ := json.Marshal(obj)
			return json.Unmarshal(b, v) == nil
		}
	}
	return false
}

// parseRow parses a payload row in `<hex id>:<data>` format and returns its
// data if it is a JSON array or object, e.g., `6:["$","main",null,{...}]`.
func parseRow(line string) (string, bool) {
	id, data, ok := strings.Cut(line, ":")
	if !ok || (!strings.HasPrefix(data, "[") && !strings.HasPrefix(data, "{")) {
		return "", false
	}
	if _, err := strconv.ParseUint(id, 16, 64); err != nil {
		return "", false
	}
	return data, true
}

func findObject(node any, keys []string) map[string]any {
	var children []any
	switch n := node.(type) {
	case map[string]any:
		if hasKeys(n, keys) {
			return n
		}
		for _, key := range slices.Sorted(maps.Keys(n)) {
			children = append(children, n[key])
		}
	case []any:
		children = n
	}
	for _, child := range children {
		if obj := findObject(child, keys); obj != nil {
			return obj
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
	return true
}
