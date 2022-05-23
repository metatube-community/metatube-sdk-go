package constant

import (
	"fmt"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
)

// VersionString returns hyphen joined version and commit string.
func VersionString() string {
	return fmt.Sprintf("%s-%s", Version, GitCommit)
}
