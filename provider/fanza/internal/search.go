package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
)

// SearchPageParser implements a search page parser for FANZA.
type SearchPageParser struct {
	vm   *otto.Otto
	data string
}

func NewSearchPageParser() *SearchPageParser {
	p := &SearchPageParser{
		vm: otto.New(),
	}
	// Pre-define var `self` to the runtime.
	_, _ = p.vm.Run("var self = {};")
	return p
}

func (p *SearchPageParser) LoadJSCode(code string) error {
	_, err := p.vm.Run(code)
	return err
}

func (p *SearchPageParser) process() error {
	self, err := p.vm.Get("self")
	if err != nil {
		return fmt.Errorf("get var `self` from runtime: %w", err)
	}

	if !self.IsObject() {
		return fmt.Errorf("var `self` from runtime is not an object: %#v", self)
	}

	value, err := self.Export()
	if err != nil {
		return fmt.Errorf("export var `self` from runtime: %w", err)
	}

	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid type of exported var `self`: %#v", value)
	}

	if _, found := object["__next_f"]; !found {
		return fmt.Errorf("var `self` does not contain key `__next_f`: %#v", object)
	}

	nextF, ok := object["__next_f"].([]any)
	if !ok {
		return fmt.Errorf("var `self.__next_f` is not an array`: %#v", object["__next_f"])
	}

	sb := &strings.Builder{}

	// extract data
	for _, v := range nextF {
		if _, ok = v.([]any); !ok {
			continue // skip invalid types
		}
		vv := v.([]any)
		if len(vv) != 2 {
			continue // skip unmatched length
		}
		idx, ok1 := vv[0].(int64)
		val, ok2 := vv[1].(string)
		if !ok1 || !ok2 {
			continue // skip unmatched pairs
		}
		if idx == 1 /* targeted index */ {
			sb.WriteString(val)
		}
	}

	p.data = sb.String()
	return nil
}

func (p *SearchPageParser) findBackendResponse(data []byte) ([]byte, error) {
	var v []any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	var find func(a []any, depth int) map[string]any
	find = func(a []any, depth int) map[string]any {
		if depth > 1e2 {
			return nil
		}
		for _, i := range a {
			if am, ok := i.(map[string]any); ok {
				if _, found := am["backendResponse"]; found {
					return am
				}
			}
			if aa, ok := i.([]any); ok {
				if m := find(aa, depth+1); len(m) > 0 {
					return m
				}
			}
		}
		return nil
	}

	m := find(v, 0)
	return json.Marshal(m)
}

func (p *SearchPageParser) Parse(resp *ResponseWrapper) error {
	// pre-process next.js data content
	if err := p.process(); err != nil {
		return err
	}

	reader := strings.NewReader(p.data)
	scanner := bufio.NewScanner(reader)

	extractor := regexp.MustCompile(`^\w:(\[.+])`)
	var jsonContent []byte

	for scanner.Scan() {
		if line := scanner.Bytes(); bytes.Contains(line, []byte("backendResponse")) && extractor.Match(line) {
			if ss := extractor.FindSubmatch(line); len(ss) == 2 {
				var err error
				if jsonContent, err = p.findBackendResponse(ss[1]); err != nil {
					return err
				}
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return json.Unmarshal(jsonContent, &resp)
}
