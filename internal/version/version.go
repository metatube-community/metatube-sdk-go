package version

import (
	"fmt"
)

// Unknown is the default value for Version or GitCommit
// when its value is unknown.
const Unknown = "unknown"

var (
	Version   = Unknown
	GitCommit = Unknown
)

func init() {
	if Version == Unknown {
		Version = modVersion()
	}
}

// modVersion returns the module version from go.mod
// when this package is used as a dependency.
func modVersion() string {
	const module = "github.com/metatube-community/metatube-sdk-go"
	for _, mod := range Modules() {
		if mod.Path == module {
			return mod.Version
		}
	}
	return Unknown
}

// BuildString returns a hyphen-joined version and commit string.
func BuildString() string {
	if GitCommit == Unknown {
		return Version
	}
	return fmt.Sprintf("v%s-%s", Version, GitCommit)
}
