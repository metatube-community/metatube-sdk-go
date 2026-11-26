// Package nextjs provides helpers to extract data from pages rendered by
// the Next.js App Router, which streams the React Server Components (RSC)
// payload to the browser via `self.__next_f.push([1, "..."])` scripts.
package nextjs

import (
	"bytes"
	"cmp"
	"encoding/base64"
	"encoding/json"
	"errors"
	"iter"
	"maps"
	"reflect"
	"regexp"
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

// Types of the chunks pushed to `self.__next_f`.
const (
	chunkData   = 1 // [1, "<payload text>"]
	chunkBinary = 3 // [3, "<base64 payload bytes>"]
)

// rowHeaderRe matches the header of a payload row: its hex id, followed by
// the tag and the hex byte length if the row data is length-prefixed instead
// of newline-terminated, e.g., `5:T1a,` for text rows.
var rowHeaderRe = regexp.MustCompile(`^[0-9a-f]+:(?:[TAOoUSsLlGgMmVb]([0-9a-f]+),)?`)

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
	for _, script := range htmlquery.Find(doc, `//script[not(@src)][contains(., "self.__next_f")]`) {
		// Skip scripts that fail, e.g., ones that also use browser APIs
		// which are not available in the VM.
		_, _ = vm.Run(htmlquery.InnerText(script))
	}

	v, err := vm.Run(`JSON.stringify(self.__next_f)`)
	if err != nil {
		return "", err
	}
	var chunks []any
	if err = json.Unmarshal([]byte(v.String()), &chunks); err != nil {
		return "", err
	}

	var b strings.Builder
	for _, chunk := range chunks {
		c, ok := chunk.([]any)
		if !ok || len(c) != 2 {
			continue
		}
		data, ok := c[1].(string)
		if !ok {
			continue
		}
		switch c[0] {
		case float64(chunkData):
			b.WriteString(data)
		case float64(chunkBinary):
			raw, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				return "", err
			}
			b.Write(raw)
		}
	}
	return b.String(), nil
}

// FindObject unmarshals the object in the RSC payload that best matches v
// into v, which must be a non-nil pointer. Objects score one point for each
// field of v they have, by its JSON name and case-insensitively; the highest
// score wins, ties go to the first one found, and objects that cannot be
// unmarshaled into v are skipped. Objects must also have all the given keys,
// which are required when v has no struct fields, e.g., a map. It returns
// false if no object matches, in which case v is left untouched.
func FindObject(payload string, v any, keys ...string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false
	}
	fields := jsonFieldNames(rv.Type().Elem())

	// A struct needs at least one matching field, while a map, which has no
	// fields, needs keys.
	minScore := 1
	if len(fields) == 0 {
		if len(keys) == 0 {
			return false
		}
		minScore = 0
	}

	type candidate struct {
		obj   map[string]any
		score int
	}
	var candidates []candidate
	for data := range jsonRows(payload) {
		var row any
		if json.Unmarshal([]byte(data), &row) != nil {
			continue
		}
		walkObjects(row, func(obj map[string]any) {
			if !hasKeys(obj, keys) {
				return
			}
			if score := matchFields(obj, fields); score >= minScore {
				candidates = append(candidates, candidate{obj, score})
			}
		})
	}

	// Try the best candidates first, keeping the payload order for ties.
	slices.SortStableFunc(candidates, func(a, b candidate) int {
		return cmp.Compare(b.score, a.score)
	})
	for _, c := range candidates {
		if decodeObject(c.obj, rv) {
			return true
		}
	}
	return false
}

// jsonRows yields the data of the payload rows that are JSON arrays or
// objects, e.g., `["$","main",null,{...}]` of the row `6:["$","main",...]`.
func jsonRows(payload string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for payload != "" {
			m := rowHeaderRe.FindStringSubmatchIndex(payload)
			if m == nil {
				// Skip a malformed line.
				_, payload, _ = strings.Cut(payload, "\n")
				continue
			}
			if m[2] >= 0 {
				// Skip length-prefixed data, which is not newline-terminated.
				n, err := strconv.ParseUint(payload[m[2]:m[3]], 16, 64)
				if err != nil || n > uint64(len(payload)-m[1]) {
					return
				}
				payload = payload[m[1]+int(n):]
				continue
			}
			var data string
			data, payload, _ = strings.Cut(payload[m[1]:], "\n")
			if strings.HasPrefix(data, "[") || strings.HasPrefix(data, "{") {
				if !yield(data) {
					return
				}
			}
		}
	}
}

// walkObjects calls visit for every object in node, depth-first, with map
// values in sorted key order.
func walkObjects(node any, visit func(map[string]any)) {
	var children []any
	switch n := node.(type) {
	case map[string]any:
		visit(n)
		for _, key := range slices.Sorted(maps.Keys(n)) {
			children = append(children, n[key])
		}
	case []any:
		children = n
	}
	for _, child := range children {
		walkObjects(child, visit)
	}
}

func hasKeys(obj map[string]any, keys []string) bool {
	for _, key := range keys {
		if _, ok := obj[key]; !ok {
			return false
		}
	}
	return true
}

// matchFields counts the fields that obj has, case-insensitively. Keys with
// undefined values do not count.
func matchFields(obj map[string]any, fields map[string]struct{}) int {
	matched := make(map[string]struct{})
	for key, value := range obj {
		if value == "$undefined" {
			continue
		}
		name := strings.ToLower(key)
		if _, ok := fields[name]; ok {
			matched[name] = struct{}{}
		}
	}
	return len(matched)
}

// decodeObject unmarshals obj into the value rv points to, which is only
// overwritten if obj is unmarshaled successfully.
func decodeObject(obj map[string]any, rv reflect.Value) bool {
	b, err := json.Marshal(decodeValue(obj))
	if err != nil {
		return false
	}
	v := reflect.New(rv.Elem().Type())
	if json.Unmarshal(b, v.Interface()) != nil {
		return false
	}
	rv.Elem().Set(v.Elem())
	return true
}

// decodeValue returns a copy of v with the special `$`-prefixed strings of
// the RSC payload decoded. References to other rows are kept as they are.
func decodeValue(v any) any {
	switch v := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(v))
		for key, value := range v {
			m[key] = decodeValue(value)
		}
		return m
	case []any:
		s := make([]any, len(v))
		for i, value := range v {
			s[i] = decodeValue(value)
		}
		return s
	case string:
		switch {
		case strings.HasPrefix(v, "$$"): // escaped `$`
			return v[1:]
		case v == "$undefined", v == "$NaN", v == "$Infinity", v == "$-Infinity":
			return nil
		case v == "$-0":
			return 0
		case strings.HasPrefix(v, "$D"): // date
			return v[2:]
		case strings.HasPrefix(v, "$n"): // bigint
			return json.Number(v[2:])
		}
	}
	return v
}

// jsonFieldNames returns the lowercased JSON names of the fields of t if it
// is a struct, following the `encoding/json` naming rules.
func jsonFieldNames(t reflect.Type) map[string]struct{} {
	if t.Kind() != reflect.Struct {
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
				// Unexported embedded struct pointers cannot be set.
				if !field.IsExported() {
					continue
				}
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
