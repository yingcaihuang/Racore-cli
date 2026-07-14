package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"racore-cli/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Error: unexpected panic: %v\n", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()
	cmd.Execute()
}
