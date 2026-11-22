// Package main is the standalone panel binary: a multi-panel TUI with
// choose and filter panels side by side.
package main

import (
	"os"
	"runtime/debug"

	"github.com/oSoWoSo/panel/v2/panel"
)

var (
	// Version contains the application version number. It's set via ldflags
	// when building.
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags when building.
	CommitSHA = ""
)

func main() {
	if Version == "" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
			Version = info.Main.Version
		} else {
			Version = "unknown (built from source)"
		}
	}
	os.Exit(panel.RunProgram(os.Args[1:], panel.VersionString(Version, CommitSHA)))
}
