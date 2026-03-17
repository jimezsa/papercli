package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jimezsa/papercli/internal/cmd"
	"github.com/jimezsa/papercli/internal/config"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fatal(err)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	globals, showVersion, showHelp, command, commandArgs, err := cmd.ParseGlobalArgs(args)
	if err != nil {
		return err
	}
	if shouldAutoInitConfig(showVersion, showHelp, command, commandArgs) {
		if _, err := config.EnsureFile(); err != nil {
			return err
		}
	}

	if len(args) == 0 {
		return cmd.PrintHelp(stdout, "", nil, globals)
	}
	if showVersion {
		_, err := fmt.Fprintln(stdout, buildVersion())
		return err
	}
	if showHelp {
		return cmd.PrintHelp(stdout, command, commandArgs, globals)
	}

	app, err := cmd.NewApp(buildVersion(), globals, stdout, stderr)
	if err != nil {
		return err
	}
	if command == "" {
		return fmt.Errorf("missing command")
	}
	return cmd.Dispatch(app, command, commandArgs)
}

func shouldAutoInitConfig(showVersion, showHelp bool, command string, args []string) bool {
	if showVersion || showHelp {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(command), "version") {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(command), "config") && len(args) > 0 {
		return !strings.EqualFold(strings.TrimSpace(args[0]), "init")
	}
	return true
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
