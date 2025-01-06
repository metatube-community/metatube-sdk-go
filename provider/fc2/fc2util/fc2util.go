package fc2util

import (
	"regexp"
)

var _fc2pattern = regexp.MustCompile(`^(?i)(?:FC2(?:[-_]?PPV)?[-_]?)?(\d+)$`)

func ParseNumber(s string) string {
	ss := _fc2pattern.FindStringSubmatch(s)
	if len(ss) != 2 {
		return ""
	}
	return ss[1]
}
