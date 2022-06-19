package version

import (
	"fmt"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
)

// VersionString returns hyphen joined version and commit string.
func VersionString() string {
	return fmt.Sprintf("v%s-%s", Version, GitCommit)
}
