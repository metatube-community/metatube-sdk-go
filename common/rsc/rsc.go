// Package rsc provides helpers for the React Server Components (RSC) payload.
package rsc

import (
	"bytes"
	"cmp"
	"context"
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

var (
	ErrNotFound  = errors.New("rsc: object not found")
	ErrAmbiguous = errors.New("rsc: multiple objects match")
)

// FindObject unmarshals the object in the RSC payload that matches the most
// fields of v into v, or fails with ErrAmbiguous if tied objects differ.
func FindObject(payload string, v any, opts ...Option) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("rsc: FindObject(non-pointer or nil %T)", v)
	}

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	fields := jsonFieldNames(v)
	if len(fields) == 0 && len(o.keys) == 0 {
		return fmt.Errorf("rsc: no fields or keys to match for %T", v)
	}
	bestScore := min(len(fields), 1) // struct objects must match a field

	var candidates []map[string]any
	for data := range jsonRows(payload) {
		var row any
		if json.Unmarshal([]byte(data), &row) != nil {
			continue
		}
		walkObjects(row, func(obj map[string]any) {
			if !hasKeys(obj, o.keys) {
				return
			}
			switch score := matchFields(obj, fields); {
			case score > bestScore:
				candidates, bestScore = []map[string]any{obj}, score
			case score == bestScore:
				candidates = append(candidates, obj)
			}
		})
	}
	if len(candidates) == 0 {
		return fmt.Errorf("%w: %T", ErrNotFound, v)
	}

	var (
		match     []byte
		first     any
		decodeErr error
	)
	t := rv.Elem().Type()
	for _, c := range candidates {
		// Decode into a zero value to compare candidates and leave v untouched.
		d := reflect.New(t).Interface()
		b, err := json.Marshal(decodeValue(c))
		if err == nil {
			err = json.Unmarshal(b, d)
		}
		if err != nil { // not a match
			decodeErr = cmp.Or(decodeErr, err)
			continue
		}
		switch {
		case first == nil:
			match, first = b, d
		case !reflect.DeepEqual(first, d):
			return fmt.Errorf("%w: %T", ErrAmbiguous, v)
		}
	}
	if first == nil {
		return decodeErr
	}
	return json.Unmarshal(match, v)
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

// walkObjects calls visit for every object in node, visiting keys in sorted order.
func walkObjects(node any, visit func(map[string]any)) {
	switch n := node.(type) {
	case map[string]any:
		visit(n)
		for _, key := range slices.Sorted(maps.Keys(n)) {
			walkObjects(n[key], visit)
		}
	case []any:
		for _, child := range n {
			walkObjects(child, visit)
		}
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
