package version

import (
	"fmt"
	"runtime/debug"
)

// Unknown is the default value for Version or GitCommit
// when its value is unknown.
const Unknown = "unknown"

var (
	Version   = Unknown
	GitCommit = Unknown
)

// version is helpful to get the version info from the
// go.mod when using this pkg as a third-party module.
func version() string {
	const (
		module = "github.com/metatube-community/metatube-sdk-go"
	)
	bi, _ := debug.ReadBuildInfo()
	for _, mod := range bi.Deps {
		if mod.Path == module {
			return mod.Version
		}
	}
	return Unknown
}

func init() {
	if Version == Unknown {
		Version = version()
	}
}

// BuildString returns hyphen joined version and commit string.
func BuildString() string {
	if GitCommit == Unknown {
		return Version
	}
	return fmt.Sprintf("v%s-%s", Version, GitCommit)
}
