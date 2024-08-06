package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
)

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
	return "unknown"
}

func init() {
	if Version == "unknown" {
		Version = version()
	}
}

// BuildString returns hyphen joined version and commit string.
func BuildString() string {
	return fmt.Sprintf("v%s-%s", Version, GitCommit)
}
