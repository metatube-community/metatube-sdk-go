// Package nextjs provides helpers to extract data from pages rendered by
// the Next.js App Router, which streams the React Server Components (RSC)
// payload to the browser via `self.__next_f.push([1, "..."])` scripts.
package nextjs

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/robertkrimen/otto"
)

// scriptTimeout limits the total execution time of the page scripts.
var scriptTimeout = 3 * time.Second

var errScriptTimeout = errors.New("nextjs: script execution timed out")

// jsonRowRe matches a payload row in `<hex id>:<data>` format whose data is
// a JSON array or object, e.g., `6:["$","main",null,{...}]`.
var jsonRowRe = regexp.MustCompile(`^[0-9a-f]+:([\[{].*)$`)

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

// FindObject unmarshals the object in the RSC payload that best matches v
// into v. Objects score one point for each field of v they have, by its JSON
// name and case-insensitively; the highest score wins, and ties go to the
// first one found. Objects must also have all the given keys, which are
// required when v has no struct fields, e.g., a map. It returns false if no
// object matches.
func FindObject(payload string, v any, keys ...string) bool {
	fields := jsonFieldNames(v)

	// The score to beat: a struct needs at least one matching field, while
	// a map, which has no fields, needs keys.
	bestScore := 0
	if len(fields) == 0 {
		if len(keys) == 0 {
			return false
		}
		bestScore = -1
	}

	var best map[string]any
	for line := range strings.SplitSeq(payload, "\n") {
		sub := jsonRowRe.FindStringSubmatch(line)
		if sub == nil {
			continue
		}
		var row any
		if json.Unmarshal([]byte(sub[1]), &row) != nil {
			continue
		}
		if walkObjects(row, func(obj map[string]any) bool {
			if !hasKeys(obj, keys) {
				return false
			}
			score := matchFields(obj, fields)
			if score <= bestScore {
				return false
			}
			best, bestScore = obj, score
			// A perfect match cannot be beaten.
			return score == len(fields)
		}) {
			break
		}
	}
	if best == nil {
		return false
	}
	// Marshaling a value decoded from JSON never fails.
	b, _ := json.Marshal(best)
	return json.Unmarshal(b, v) == nil
}

// walkObjects calls visit for every object in node, depth-first, until visit
// returns true. It reports whether visit returned true.
func walkObjects(node any, visit func(map[string]any) bool) bool {
	var children []any
	switch n := node.(type) {
	case map[string]any:
		if visit(n) {
			return true
		}
		for _, key := range slices.Sorted(maps.Keys(n)) {
			children = append(children, n[key])
		}
	case []any:
		children = n
	}
	for _, child := range children {
		if walkObjects(child, visit) {
			return true
		}
	}
	return false
}

func hasKeys(obj map[string]any, keys []string) bool {
	for _, key := range keys {
		if _, ok := obj[key]; !ok {
			return false
		}
	}
	return true
}

// matchFields counts the keys of obj that are in fields, case-insensitively.
func matchFields(obj map[string]any, fields map[string]struct{}) (n int) {
	for key := range obj {
		if _, ok := fields[strings.ToLower(key)]; ok {
			n++
		}
	}
	return
}

// jsonFieldNames returns the lowercased JSON names of the fields of v if it
// is a struct or a pointer to one, following the `encoding/json` rules.
func jsonFieldNames(v any) map[string]struct{} {
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	names := make(map[string]struct{})
	addJSONFieldNames(t, names)
	return names
}

func addJSONFieldNames(t reflect.Type, names map[string]struct{}) {
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		// Fields of untagged embedded structs are promoted.
		if field.Anonymous && name == "" {
			ft := field.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				addJSONFieldNames(ft, names)
				continue
			}
		}
		if !field.IsExported() {
			continue
		}
		if name == "" {
			name = field.Name
		}
		names[strings.ToLower(name)] = struct{}{}
	}
}
