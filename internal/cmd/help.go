package cmd

import (
	"fmt"
	"io"
	"strings"
)

func PrintHelp(w io.Writer, command string, args []string) error {
	command = strings.ToLower(strings.TrimSpace(command))
	switch command {
	case "":
		printGlobalHelp(w)
		return nil
	case "version":
		_, _ = fmt.Fprintln(w, "Usage:")
		_, _ = fmt.Fprintln(w, "  papercli version")
		return nil
	case "config":
		printConfigHelp(w)
		return nil
	case "search":
		printSearchHelp(w)
		return nil
	case "author":
		printAuthorHelp(w)
		return nil
	case "info":
		printInfoHelp(w)
		return nil
	case "download":
		printDownloadHelp(w)
		return nil
	case "seen":
		if len(args) > 0 {
			switch strings.ToLower(strings.TrimSpace(args[0])) {
			case "diff":
				printSeenDiffHelp(w)
				return nil
			case "update":
				printSeenUpdateHelp(w)
				return nil
			}
		}
		printSeenHelp(w)
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func printGlobalHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "papercli - search and aggregate academic papers")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli [global flags] <command> [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Global flags:")
	_, _ = fmt.Fprintln(w, "  --color auto|always|never")
	_, _ = fmt.Fprintln(w, "  --json")
	_, _ = fmt.Fprintln(w, "  --plain")
	_, _ = fmt.Fprintln(w, "  --verbose")
	_, _ = fmt.Fprintln(w, "  --version")
	_, _ = fmt.Fprintln(w, "  --help, -h")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  version")
	_, _ = fmt.Fprintln(w, "  config init|path")
	_, _ = fmt.Fprintln(w, "  search <query>")
	_, _ = fmt.Fprintln(w, "  author <name>")
	_, _ = fmt.Fprintln(w, "  info <id>")
	_, _ = fmt.Fprintln(w, "  download <id>")
	_, _ = fmt.Fprintln(w, "  seen diff|update")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Use 'papercli <command> --help' for command-specific options.")
}

func printConfigHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli config init")
	_, _ = fmt.Fprintln(w, "  papercli config path")
}

func printSearchHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli search <query> [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --provider arxiv|semantic|scholar|all")
	_, _ = fmt.Fprintln(w, "  --sort relevance|date|citations")
	_, _ = fmt.Fprintln(w, "  --year-from <year>")
	_, _ = fmt.Fprintln(w, "  --year-to <year>")
	_, _ = fmt.Fprintln(w, "  --limit <n>")
	_, _ = fmt.Fprintln(w, "  --offset <n>")
	_, _ = fmt.Fprintln(w, "  --format csv|json|md")
	_, _ = fmt.Fprintln(w, "  --links short|full")
	_, _ = fmt.Fprintln(w, "  --seen <path>")
	_, _ = fmt.Fprintln(w, "  --new-only")
	_, _ = fmt.Fprintln(w, "  --new-out <path>")
	_, _ = fmt.Fprintln(w, "  --out, --output <path>")
}

func printAuthorHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli author <name> [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Flags are the same as 'papercli search'.")
}

func printInfoHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli info <id> [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --provider arxiv|semantic|scholar|all")
	_, _ = fmt.Fprintln(w, "  --format csv|json|md")
	_, _ = fmt.Fprintln(w, "  --links short|full")
	_, _ = fmt.Fprintln(w, "  --out, --output <path>")
}

func printDownloadHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli download <id> [flags]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  --provider arxiv|semantic|scholar|all")
	_, _ = fmt.Fprintln(w, "  --out, --output, --file <path>")
}

func printSeenHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli seen diff --new A.json --seen B.json --out C.json [--stats]")
	_, _ = fmt.Fprintln(w, "  papercli seen update --seen B.json --input C.json --out B.json [--stats]")
}

func printSeenDiffHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli seen diff --new A.json --seen B.json --out C.json [--stats]")
}

func printSeenUpdateHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  papercli seen update --seen B.json --input C.json --out B.json [--stats]")
}
