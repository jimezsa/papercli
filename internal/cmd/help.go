package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/jimezsa/papercli/internal/ui"
	"github.com/muesli/termenv"
)

type helpStyler func(string) string

func PrintHelp(w io.Writer, command string, args []string, globals Globals) error {
	style := newHelpStyler(w, globals)
	command = strings.ToLower(strings.TrimSpace(command))
	switch command {
	case "":
		printGlobalHelp(w, style)
		return nil
	case "version":
		_, _ = fmt.Fprintln(w, style("Usage:"))
		_, _ = fmt.Fprintf(w, "  %s\n", style("papercli version"))
		return nil
	case "config":
		printConfigHelp(w, style)
		return nil
	case "search":
		printSearchHelp(w, style)
		return nil
	case "author":
		printAuthorHelp(w, style)
		return nil
	case "info":
		printInfoHelp(w, style)
		return nil
	case "download":
		printDownloadHelp(w, style)
		return nil
	case "seen":
		if len(args) > 0 {
			switch strings.ToLower(strings.TrimSpace(args[0])) {
			case "diff":
				printSeenDiffHelp(w, style)
				return nil
			case "update":
				printSeenUpdateHelp(w, style)
				return nil
			}
		}
		printSeenHelp(w, style)
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func newHelpStyler(w io.Writer, globals Globals) helpStyler {
	enabled := ui.ColorsEnabled(globals.Color, globals.JSON || globals.Plain)
	output := termenv.NewOutput(w)
	return func(text string) string {
		return ui.ColorizeLink(output, enabled, text)
	}
}

func printGlobalHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("papercli - search and aggregate academic papers"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s <command> [flags]\n", style("papercli"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Global flags:"))
	_, _ = fmt.Fprintf(w, "  %s auto|always|never\n", style("--color"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("--json"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("--plain"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("--verbose"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("--version"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("--help, -h"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Global flags can be placed before or after <command>."))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Commands:"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("version"))
	_, _ = fmt.Fprintf(w, "  %s init|path\n", style("config"))
	_, _ = fmt.Fprintf(w, "  %s <query>\n", style("search"))
	_, _ = fmt.Fprintf(w, "  %s <name>\n", style("author"))
	_, _ = fmt.Fprintf(w, "  %s <id>\n", style("info"))
	_, _ = fmt.Fprintf(w, "  %s <id>\n", style("download"))
	_, _ = fmt.Fprintf(w, "  %s diff|update\n", style("seen"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Examples:"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("papercli config init"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("papercli search \"vision transformer\" --provider arxiv --limit 5"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("papercli info 1706.03762 --provider arxiv --format md"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("papercli download 1706.03762 --provider arxiv --out attention-is-all-you-need.pdf"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("papercli seen diff --new latest.json --seen seen.json --out unseen.json --stats"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Use '%s <command> %s' for command-specific options.\n", style("papercli"), style("--help"))
}

func printConfigHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s config init [%s]\n", style("papercli"), style("--force"))
	_, _ = fmt.Fprintf(w, "  %s config path\n", style("papercli"))
}

func printSearchHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s search <query> [flags]\n", style("papercli"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Flags:"))
	_, _ = fmt.Fprintf(w, "  %s arxiv|semantic|scholar|all\n", style("--provider"))
	_, _ = fmt.Fprintf(w, "  %s relevance|date|citations\n", style("--sort"))
	_, _ = fmt.Fprintf(w, "  %s <year>\n", style("--year-from"))
	_, _ = fmt.Fprintf(w, "  %s <year>\n", style("--year-to"))
	_, _ = fmt.Fprintf(w, "  %s <n>\n", style("--limit"))
	_, _ = fmt.Fprintf(w, "  %s <n>\n", style("--offset"))
	_, _ = fmt.Fprintf(w, "  %s csv|json|md\n", style("--format"))
	_, _ = fmt.Fprintf(w, "  %s short|full\n", style("--links"))
	_, _ = fmt.Fprintf(w, "  %s <path>\n", style("--seen"))
	_, _ = fmt.Fprintf(w, "  %s\n", style("--new-only"))
	_, _ = fmt.Fprintf(w, "  %s <path>\n", style("--new-out"))
	_, _ = fmt.Fprintf(w, "  %s <path>\n", style("--out, --output"))
}

func printAuthorHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s author <name> [flags]\n", style("papercli"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Flags are the same as '%s search'.\n", style("papercli"))
}

func printInfoHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s info <id> [flags]\n", style("papercli"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Flags:"))
	_, _ = fmt.Fprintf(w, "  %s arxiv|semantic|scholar|all\n", style("--provider"))
	_, _ = fmt.Fprintf(w, "  %s csv|json|md\n", style("--format"))
	_, _ = fmt.Fprintf(w, "  %s short|full\n", style("--links"))
	_, _ = fmt.Fprintf(w, "  %s <path>\n", style("--out, --output"))
}

func printDownloadHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s download <id> [flags]\n", style("papercli"))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, style("Flags:"))
	_, _ = fmt.Fprintf(w, "  %s arxiv|semantic|scholar|all\n", style("--provider"))
	_, _ = fmt.Fprintf(w, "  %s <path>\n", style("--out, --output, --file"))
}

func printSeenHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s seen diff %s A.json %s B.json %s C.json [%s]\n", style("papercli"), style("--new"), style("--seen"), style("--out"), style("--stats"))
	_, _ = fmt.Fprintf(w, "  %s seen update %s B.json %s C.json %s B.json [%s]\n", style("papercli"), style("--seen"), style("--input"), style("--out"), style("--stats"))
}

func printSeenDiffHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s seen diff %s A.json %s B.json %s C.json [%s]\n", style("papercli"), style("--new"), style("--seen"), style("--out"), style("--stats"))
}

func printSeenUpdateHelp(w io.Writer, style helpStyler) {
	_, _ = fmt.Fprintln(w, style("Usage:"))
	_, _ = fmt.Fprintf(w, "  %s seen update %s B.json %s C.json %s B.json [%s]\n", style("papercli"), style("--seen"), style("--input"), style("--out"), style("--stats"))
}
