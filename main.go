package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"racore-cli/cmd"
)

// These are set by goreleaser at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Error: unexpected panic: %v\n", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
