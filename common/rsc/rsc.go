// Package rsc provides helpers for the React Server Components (RSC) payload.
package rsc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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

	"github.com/metatube-community/metatube-sdk-go/common/js"
)

// rowHeaderRe matches a row header, capturing the length of length-prefixed rows.
var rowHeaderRe = regexp.MustCompile(`^[0-9a-f]+:(?:[TAOoUSsLlGgMmVb]([0-9a-f]+),)?`)

// ExtractNextPayload extracts the RSC payload from a Next.js App Router page.
func ExtractNextPayload(page []byte) (string, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(page))
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var data string
	if err = js.RunVM(ctx, func(vm *otto.Otto) error {
		if _, err := vm.Run(`var self = { __next_f: [] };`); err != nil {
			return err
		}
		// Other scripts may depend on browser APIs.
		for _, script := range htmlquery.Find(doc, `//script[not(@src)][contains(., "self.__next_f")]`) {
			if _, err := vm.Run(htmlquery.InnerText(script)); err != nil {
				return err
			}
		}
		v, err := vm.Run(`JSON.stringify(self.__next_f)`)
		if err != nil {
			return err
		}
		data = v.String()
		return nil
	}); err != nil {
		return "", err
	}

	var chunks []any
	if err = json.Unmarshal([]byte(data), &chunks); err != nil {
		return "", err
	}

	// Only text and binary chunks carry the payload.
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

// FindObject unmarshals the object in the RSC payload that matches the most
// fields of v into v. Objects must have all the given keys, which are required
// if v is not a struct.
func FindObject(payload string, v any, keys ...string) bool {
	fields := jsonFieldNames(v)

	bestScore := 0
	if len(fields) == 0 {
		if len(keys) == 0 {
			return false
		}
		bestScore = -1
	}

	var best map[string]any
	for data := range jsonRows(payload) {
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
	b, err := json.Marshal(decodeValue(best))
	if err != nil {
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false
	}
	// Unmarshal into a zero value first to leave v untouched on failure.
	if json.Unmarshal(b, reflect.New(rv.Elem().Type()).Interface()) != nil {
		return false
	}
	return json.Unmarshal(b, v) == nil
}

// jsonRows yields the data of the JSON rows in the payload.
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

// decodeValue returns a copy of v with the `$`-prefixed special values decoded.
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
		case strings.HasPrefix(v, "$n"): // bigint, decoded as a number
			return json.Number(v[2:])
		}
	}
	return v
}

// walkObjects calls visit for every object in node until it returns true.
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

// matchFields counts the fields that obj has, case-insensitively.
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

// jsonFieldNames returns the lowercased JSON field names of v.
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
