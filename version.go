// Package main tells the version of the edfbleu binary, read from its build: the module version
// Go stamps (a tag, or a pseudo-version from the commit, "+dirty" for uncommitted changes), with
// the Go version and platform for humans.
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Version returns the version of the binary for humans (--version): the module version, the Go
// version and the platform it was built for.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("unknown (%s/%s)", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("%s (%s, %s/%s)", info.Main.Version, info.GoVersion, runtime.GOOS, runtime.GOARCH)
}
