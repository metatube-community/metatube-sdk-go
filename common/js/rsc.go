package js

// This file provides helpers to extract data from pages rendered by the
// Next.js App Router, which streams the React Server Components (RSC)
// payload to the browser via `self.__next_f.push([1, "..."])` scripts.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

// rscScriptTimeout limits the total execution time of the page scripts.
var rscScriptTimeout = 3 * time.Second

var errRSCScriptTimeout = errors.New("rsc: script execution timed out")

// rscRowHeaderRe matches the header of a payload row: its hex id, followed by
// the tag and the hex byte length if the row data is length-prefixed instead
// of newline-terminated, e.g., `5:T1a,` for text rows.
var rscRowHeaderRe = regexp.MustCompile(`^[0-9a-f]+:(?:[TAOoUSsLlGgMmVb]([0-9a-f]+),)?`)

// ExtractRSCPayload extracts the RSC payload from a Next.js App Router HTML
// page. It runs the inline flight data scripts of the page in a JS VM with
// a predefined `self.__next_f` array, then concatenates the pushed payload
// chunks.
func ExtractRSCPayload(page []byte) (payload string, err error) {
	doc, err := htmlquery.Parse(bytes.NewReader(page))
	if err != nil {
		return "", err
	}

	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)

	timer := time.AfterFunc(rscScriptTimeout, func() {
		vm.Interrupt <- func() { panic(errRSCScriptTimeout) }
	})
	defer timer.Stop()

	defer func() {
		// The scripts come from untrusted pages, so report any panic of the
		// VM as an error, not only the timeout.
		if r := recover(); r != nil {
			payload = ""
			if r == errRSCScriptTimeout {
				err = errRSCScriptTimeout
			} else {
				err = fmt.Errorf("rsc: script panicked: %v", r)
			}
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

	v, err := vm.Run(`JSON.stringify(self.__next_f)`)
	if err != nil {
		return "", err
	}
	var chunks []any
	if err = json.Unmarshal([]byte(v.String()), &chunks); err != nil {
		return "", err
	}

	// Only text (`[1, "..."]`) and binary (`[3, "<base64>"]`) chunks carry
	// the payload data.
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
		case float64(1):
			b.WriteString(data)
		case float64(3):
			raw, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				return "", err
			}
			b.Write(raw)
		}
	}
	return b.String(), nil
}

// FindRSCObject unmarshals the object in the RSC payload that best matches v
// into v. Objects score one point for each field of v they have, by its JSON
// name and case-insensitively; the highest score wins, and ties go to the
// first one found. Objects must also have all the given keys, matched
// exactly, which are required when v has no struct fields, e.g., a map.
//
// The object is merged into v like json.Unmarshal does, keeping what v has
// that the object does not. It returns false if no object matches, or if the
// best one cannot be unmarshaled into v, which is then left untouched.
// References to other rows are kept as `$<id>` strings, so fields that may
// hold them should be typed as string or any.
func FindRSCObject(payload string, v any, keys ...string) bool {
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
	for data := range rscJSONRows(payload) {
		var row any
		if json.Unmarshal([]byte(data), &row) != nil {
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
	b, err := json.Marshal(decodeRSCValue(best))
	if err != nil {
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false
	}
	// Check that the object fits v's type on a zero value first, so v is left
	// untouched on failure, then unmarshal into v itself to merge like
	// json.Unmarshal.
	if json.Unmarshal(b, reflect.New(rv.Elem().Type()).Interface()) != nil {
		return false
	}
	return json.Unmarshal(b, v) == nil
}

// rscJSONRows yields the data of the payload rows that are JSON arrays or
// objects, e.g., `["$","main",null,{...}]` of the row `6:["$","main",...]`.
func rscJSONRows(payload string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for payload != "" {
			m := rscRowHeaderRe.FindStringSubmatchIndex(payload)
			if m == nil {
				// Skip a malformed line.
				_, payload, _ = strings.Cut(payload, "\n")
				continue
			}
			if m[2] >= 0 {
				// Skip length-prefixed data, which may contain newlines.
				n, err := strconv.ParseUint(payload[m[2]:m[3]], 16, 64)
				if err != nil || n > uint64(len(payload)-m[1]) {
					return // truncated payload
				}
				payload = payload[uint64(m[1])+n:]
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

// decodeRSCValue returns a copy of v with the simple `$`-prefixed values of the
// RSC payload decoded: escaped strings, undefined and special numbers, dates
// and bigints. Other values, e.g., references to other rows, are kept as is.
func decodeRSCValue(v any) any {
	switch v := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(v))
		for key, value := range v {
			m[key] = decodeRSCValue(value)
		}
		return m
	case []any:
		s := make([]any, len(v))
		for i, value := range v {
			s[i] = decodeRSCValue(value)
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
		case strings.HasPrefix(v, "$n"): // bigint, decoded as a number
			return json.Number(v[2:])
		}
	}
	return v
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

// matchFields counts the fields that obj has, case-insensitively. Keys that
// differ only in case count once.
func matchFields(obj map[string]any, fields map[string]struct{}) int {
	matched := make(map[string]struct{})
	for key := range obj {
		name := strings.ToLower(key)
		if _, ok := fields[name]; ok {
			matched[name] = struct{}{}
		}
	}
	return len(matched)
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
