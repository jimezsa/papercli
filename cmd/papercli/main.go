package main

import (
	"fmt"
	"os"

	"github.com/jimezsa/papercli/internal/cmd"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

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
		fmt.Println(buildVersion())
		return
	}
	if showHelp {
		if err := cmd.PrintHelp(os.Stdout, command, args, globals); err != nil {
			fatal(err)
		}
		return
	}

	app, err := cmd.NewApp(buildVersion(), globals, os.Stdout, os.Stderr)
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

func buildVersion() string {
	if commit == "" && date == "" {
		return version
	}
	if commit == "" {
		return fmt.Sprintf("%s (%s)", version, date)
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}
