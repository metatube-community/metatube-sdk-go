package version

import (
	"runtime/debug"
)

var modules []*debug.Module

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("debug.ReadBuildInfo() failure")
	}
	modules = bi.Deps
}

// Modules returns all go dependencies/modules.
func Modules() []*debug.Module {
	return modules
}
