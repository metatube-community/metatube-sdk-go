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
var rowHeaderRe = regexp.MustCompile(`^([0-9a-f]+):(?:([TAOoUSsLlGgMmVb])([0-9a-f]+),)?`)

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
	var scriptErr error
	for _, script := range htmlquery.Find(doc, `//script[not(@src)][contains(., "self.__next_f")]`) {
		// Skip scripts that fail, e.g., ones that also use browser APIs
		// which are not available in the VM.
		if _, err := vm.Run(htmlquery.InnerText(script)); err != nil && scriptErr == nil {
			scriptErr = err
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
	if b.Len() == 0 && scriptErr != nil {
		// Report the failure if no data could be extracted at all.
		return "", scriptErr
	}
	return b.String(), nil
}

// FindObject unmarshals the object in the RSC payload that best matches v
// into v, which must be a non-nil pointer. Objects score one point for each
// field of v they have, by its JSON name and case-insensitively; the highest
// score wins, ties go to the first one found, and objects that cannot be
// unmarshaled into v are skipped. Objects must also have all the given keys,
// also compared case-insensitively, which are required when v has no struct
// fields, e.g., a map. It returns false if no object matches, in which case
// v is left untouched.
//
// The special `$`-prefixed values of the payload are decoded, and references
// to other rows, e.g., long strings sent as text rows, are resolved.
func FindObject(payload string, v any, keys ...string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false
	}
	t, ok := indirectType(rv.Type().Elem())
	if !ok {
		return false
	}
	fields := jsonFieldNames(t)

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
	rows := parseRows(payload)
	for _, row := range rows.values {
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
		if decodeObject(c.obj, rows.byID, rv) {
			return true
		}
	}
	return false
}

// payloadRows holds the parsed rows of an RSC payload.
type payloadRows struct {
	byID   map[string]any // JSON values and text rows by row id.
	values []any          // JSON arrays and objects in payload order.
}

// parseRows parses the rows of an RSC payload. Rows are newline-terminated,
// except length-prefixed ones, e.g., `5:T1a,<text>` for text rows.
func parseRows(payload string) payloadRows {
	rows := payloadRows{byID: make(map[string]any)}
	for payload != "" {
		m := rowHeaderRe.FindStringSubmatchIndex(payload)
		if m == nil {
			// Skip a malformed line.
			_, payload, _ = strings.Cut(payload, "\n")
			continue
		}
		id := payload[m[2]:m[3]]

		if m[4] >= 0 {
			n, err := strconv.ParseUint(payload[m[6]:m[7]], 16, 64)
			if err != nil || n > uint64(len(payload)-m[1]) {
				break // truncated payload
			}
			end := uint64(m[1]) + n
			if payload[m[4]:m[5]] == "T" {
				rows.byID[id] = payload[m[1]:end]
			}
			payload = payload[end:]
			continue
		}

		var data string
		data, payload, _ = strings.Cut(payload[m[1]:], "\n")
		var value any
		if json.Unmarshal([]byte(data), &value) != nil {
			continue // e.g., module or hint rows
		}
		rows.byID[id] = value
		switch value.(type) {
		case map[string]any, []any:
			rows.values = append(rows.values, value)
		}
	}
	return rows
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

// hasKeys reports whether obj has all the keys, case-insensitively.
func hasKeys(obj map[string]any, keys []string) bool {
	for _, key := range keys {
		if !hasKey(obj, key) {
			return false
		}
	}
	return true
}

func hasKey(obj map[string]any, key string) bool {
	if _, ok := obj[key]; ok {
		return true
	}
	for k := range obj {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
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
// overwritten if obj is unmarshaled successfully. Bigints are decoded into
// numbers, or into strings if v needs them to be strings.
func decodeObject(obj map[string]any, rows map[string]any, rv reflect.Value) bool {
	for _, bigintAsString := range []bool{false, true} {
		d := newDecoder(rows, bigintAsString)
		decoded, _ := d.decode(obj)
		b, err := json.Marshal(decoded)
		if err == nil {
			v := reflect.New(rv.Elem().Type())
			if json.Unmarshal(b, v.Interface()) == nil {
				rv.Elem().Set(v.Elem())
				return true
			}
		}
		if !d.sawBigint {
			break // retrying with string bigints would not change anything
		}
	}
	return false
}

// maxRefHops limits how many references are followed to reach a value.
const maxRefHops = 32

// maxDecodedSize limits the size of a decoded value, counted in JSON values.
// Shared references are expanded into copies, which could otherwise grow
// exponentially; references beyond the limit are kept as they are.
const maxDecodedSize = 1 << 18

// decoder decodes the special `$`-prefixed values of an RSC payload.
type decoder struct {
	rows           map[string]any
	bigintAsString bool
	sawBigint      bool

	// Each reference is decoded once, and references that are still being
	// decoded, i.e., cyclic ones, are kept as they are.
	decoded   map[string]decodedValue
	resolving map[string]bool
}

type decodedValue struct {
	value any
	size  int // number of JSON values once expanded
}

func newDecoder(rows map[string]any, bigintAsString bool) *decoder {
	return &decoder{
		rows:           rows,
		bigintAsString: bigintAsString,
		decoded:        make(map[string]decodedValue),
		resolving:      make(map[string]bool),
	}
}

// decode returns a copy of v with the special values decoded, and its size.
func (d *decoder) decode(v any) (any, int) {
	switch v := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(v))
		size := 1
		for _, key := range slices.Sorted(maps.Keys(v)) {
			var n int
			m[key], n = d.decodeChild(v[key], size)
			size += n
		}
		return m, size
	case []any:
		s := make([]any, len(v))
		size := 1
		for i, value := range v {
			var n int
			s[i], n = d.decodeChild(value, size)
			size += n
		}
		return s, size
	case string:
		return d.decodeString(v)
	}
	return v, 1
}

// decodeChild decodes a child of a value whose size is size so far, keeping
// a reference as it is if expanding it would exceed maxDecodedSize.
func (d *decoder) decodeChild(v any, size int) (any, int) {
	decoded, n := d.decode(v)
	if s, isString := v.(string); isString && size+n > maxDecodedSize {
		return s, 1
	}
	return decoded, n
}

func (d *decoder) decodeString(s string) (any, int) {
	switch {
	case !strings.HasPrefix(s, "$"):
		return s, 1
	case strings.HasPrefix(s, "$$"): // escaped `$`
		return s[1:], 1
	case s == "$undefined", s == "$NaN", s == "$Infinity", s == "$-Infinity":
		return nil, 1
	case s == "$-0":
		return 0, 1
	case strings.HasPrefix(s, "$D"): // date
		return s[2:], 1
	case strings.HasPrefix(s, "$n"): // bigint
		d.sawBigint = true
		if d.bigintAsString {
			return s[2:], 1
		}
		return json.Number(s[2:]), 1
	}

	if v, ok := d.decoded[s]; ok {
		return v.value, v.size
	}
	if d.resolving[s] {
		return s, 1
	}
	value, ok := d.resolve(s, 0)
	if !ok {
		return s, 1
	}
	d.resolving[s] = true
	value, size := d.decode(value)
	delete(d.resolving, s)
	d.decoded[s] = decodedValue{value, size}
	return value, size
}

// resolve resolves a reference to a row, `$<hex id>[:path]`, also as a
// promise (`$@`) or a lazy value (`$L`). The path walks into the row value,
// e.g., `$6:props:children:0`, following references on the way.
func (d *decoder) resolve(ref string, hops int) (any, bool) {
	ref, ok := strings.CutPrefix(ref, "$")
	if !ok || hops >= maxRefHops {
		return nil, false
	}
	ref = strings.TrimPrefix(strings.TrimPrefix(ref, "@"), "L")
	if ref == "" || !isHexDigit(ref[0]) {
		return nil, false
	}

	id, path, _ := strings.Cut(ref, ":")
	value, ok := d.rows[id]
	if !ok || path == "" {
		return value, ok
	}
	for key := range strings.SplitSeq(path, ":") {
		// Path steps may go through other references.
		for {
			s, isString := value.(string)
			if !isString {
				break
			}
			hops++
			if value, ok = d.resolve(s, hops); !ok {
				return nil, false
			}
		}
		switch n := value.(type) {
		case map[string]any:
			if value, ok = n[key]; !ok {
				return nil, false
			}
		case []any:
			// React elements are encoded as ["$", type, key, props].
			if len(n) >= 4 && n[0] == "$" {
				switch key {
				case "type":
					value = n[1]
				case "key":
					value = n[2]
				case "props":
					value = n[3]
				default:
					return nil, false
				}
				continue
			}
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(n) {
				return nil, false
			}
			value = n[i]
		default:
			return nil, false
		}
	}
	return value, true
}

func isHexDigit(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f'
}

// indirectType returns the type t points to through any pointers, or false
// if the pointers never end, e.g., for `type P *P`.
func indirectType(t reflect.Type) (reflect.Type, bool) {
	seen := make(map[reflect.Type]bool)
	for t.Kind() == reflect.Pointer {
		if seen[t] {
			return nil, false
		}
		seen[t] = true
		t = t.Elem()
	}
	return t, true
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
