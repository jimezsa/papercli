package main

import (
	"fmt"
	"os"

	"papercli/internal/cmd"
)

var version = "0.1.0"

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: papercli [global flags] <command> [flags]")
		os.Exit(1)
	}

	globals, showVersion, command, args, err := cmd.ParseGlobalArgs(os.Args[1:])
	if err != nil {
		fatal(err)
	}
	if showVersion {
		fmt.Println(version)
		return
	}

	app, err := cmd.NewApp(version, globals, os.Stdout, os.Stderr)
	if err != nil {
		fatal(err)
	}
	if command == "" {
		fatal(fmt.Errorf("missing command"))
	}
	if err := cmd.Dispatch(app, command, args); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
