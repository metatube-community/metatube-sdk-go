package testkit

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// getFrame: https://stackoverflow.com/a/35213181/9243111
func getFrame(skipFrames int) runtime.Frame {
	targetFrameIndex := skipFrames + 2
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)
	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}
	return frame
}

func getStructFieldByName(s any, n string) (any, error) {
	if strings.TrimSpace(n) == "" {
		return nil, errors.New("field name cannot be empty")
	}
	v := reflect.ValueOf(s)
	switch v.Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		return getStructFieldByName(v.Elem().Interface(), n)
	default:
		return nil, fmt.Errorf("wrong type: %s is a %s", v.Type(), v.Kind())
	}
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid value of %s: %s", v.Type(), v)
	}
	for i, t := 0, v.Type(); i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.EqualFold(field.Name, n) ||
			strings.EqualFold(field.Tag.Get("json"), n) {
			return v.Field(i).Interface(), nil
		}
	}
	return nil, fmt.Errorf("%s: field '%s' doesn't exist", v.Type(), n)
}

var _functionParser = regexp.MustCompile(`\.Test([^\W_]+)_([^\W_]+)`)

func parseTestFunction(function string) (string, string, error) {
	results := _functionParser.FindStringSubmatch(function)
	if len(results) != 3 {
		return "", "", fmt.Errorf("invalid test function name: %s", function)
	}
	return results[1] /* provider struct name */, results[2] /* test function name */, nil
}
