package main

import (
	"fmt"
	"github.com/jimezsa/papercli/internal/cmd"
	"os"
)

var version = "0.1.0"

func main() {
	if len(os.Args) == 1 {
		globals, _, _, _, _, _ := cmd.ParseGlobalArgs(nil)
		_ = cmd.PrintHelp(os.Stdout, "", nil, globals)
		return
	}

	globals, showVersion, showHelp, command, args, err := cmd.ParseGlobalArgs(os.Args[1:])
	if err != nil {
		fatal(err)
	}
	if showVersion {
		fmt.Println(version)
		return
	}
	if showHelp {
		if err := cmd.PrintHelp(os.Stdout, command, args, globals); err != nil {
			fatal(err)
		}
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
