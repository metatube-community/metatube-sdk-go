package version

import (
	"fmt"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
)

// BuildString returns hyphen joined version and commit string.
func BuildString() string {
	return fmt.Sprintf("v%s-%s", Version, GitCommit)
}
