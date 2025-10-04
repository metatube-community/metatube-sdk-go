package fc2util

import (
	_ "embed"
	"regexp"
)

var fc2pattern = regexp.MustCompile(`^(?i)(?:FC2(?:[-_\s]?PPV)?[-_\s]?)?(\d+)$`)

func ParseNumber(id string) string {
	ss := fc2pattern.FindStringSubmatch(id)
	if len(ss) != 2 {
		return ""
	}
	return ss[1]
}
