package fc2util

import (
	"regexp"
)

var _fc2pattern = regexp.MustCompile(`^(?i)(?:FC2(?:[-_]?PPV)?[-_]?)?(\d+)$`)

func ParseNumber(s string) (n string) {
	if ss := _fc2pattern.FindStringSubmatch(s); len(ss) == 2 {
		n = ss[1]
	}
	return
}
