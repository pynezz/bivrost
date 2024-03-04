package version

// path: pkg/version/version.go

import (
	"fmt"
	"runtime"
)

// Info returns version information
func Info() string {
	return fmt.Sprintf("Version: %s\nGit commit: %s\nGo version: %s\nOS/Arch: %s/%s\nBuild date: %s\n", version, commit, runtime.Version(), runtime.GOOS, runtime.GOARCH, buildDate)
}

var (
	version   = "dev"
	commit    = "none"
	buildDate = "na"
)

// SetVersion sets the version information
func SetVersion(v, c, b string) {
	version = v
	commit = c
	buildDate = b
}

// Version returns the version
func Version() string {
	return version
}
