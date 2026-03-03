package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseGlobalArgs(args []string) (Globals, bool, bool, string, []string, error) {
	globals := Globals{
		Color:   envString("PAPERCLI_COLOR", "auto"),
		JSON:    envBool("PAPERCLI_JSON", false),
		Verbose: envBool("PAPERCLI_VERBOSE", false),
	}

	var showVersion bool
	var showHelp bool
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		token := args[i]
		if token == "--" {
			rest = append(rest, args[i+1:]...)
			break
		}

		switch {
		case token == "--json":
			globals.JSON = true
		case token == "--plain":
			globals.Plain = true
		case token == "--verbose":
			globals.Verbose = true
		case token == "--version":
			showVersion = true
		case token == "--help" || token == "-h":
			showHelp = true
		case token == "--color":
			if i+1 >= len(args) {
				return Globals{}, false, false, "", nil, errors.New("missing value for --color")
			}
			i++
			globals.Color = strings.TrimSpace(args[i])
		case strings.HasPrefix(token, "--color="):
			globals.Color = strings.TrimSpace(strings.TrimPrefix(token, "--color="))
		default:
			rest = append(rest, token)
		}
	}

	if err := validateColorMode(globals.Color); err != nil {
		return Globals{}, false, false, "", nil, err
	}
	if err := validateOutputMode(globals); err != nil {
		return Globals{}, false, false, "", nil, err
	}

	if showVersion {
		return globals, true, showHelp, "version", nil, nil
	}
	if len(rest) > 0 && strings.HasPrefix(rest[0], "-") {
		return Globals{}, false, false, "", nil, fmt.Errorf("unknown global flag %q", rest[0])
	}
	if len(rest) == 0 {
		return globals, false, showHelp, "", nil, nil
	}
	return globals, false, showHelp, rest[0], rest[1:], nil
}

func Dispatch(app *App, command string, args []string) error {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "version":
		return (&VersionCmd{}).Run(app)
	case "config":
		return dispatchConfig(app, args)
	case "search":
		return dispatchSearch(app, args)
	case "author":
		return dispatchAuthor(app, args)
	case "info":
		return dispatchInfo(app, args)
	case "download":
		return dispatchDownload(app, args)
	case "seen":
		return dispatchSeen(app, args)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func dispatchConfig(app *App, args []string) error {
	if hasHelpFlag(args) {
		printConfigHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}
	if len(args) == 0 {
		return errors.New("config requires a subcommand: init | path")
	}
	switch args[0] {
	case "init":
		cmd := InitConfigCmd{}
		fs := flag.NewFlagSet("config init", flag.ContinueOnError)
		fs.SetOutput(app.Stderr)
		fs.BoolVar(&cmd.Force, "force", cmd.Force, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if len(fs.Args()) > 0 {
			return errors.New("config init does not accept positional arguments")
		}
		return cmd.Run(app)
	case "path":
		if len(args) > 1 {
			return errors.New("config path does not accept arguments")
		}
		return (&PathConfigCmd{}).Run(app)
	default:
		return fmt.Errorf("unknown config subcommand %q", args[0])
	}
}

func dispatchSearch(app *App, args []string) error {
	if hasHelpFlag(args) {
		printSearchHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}

	cmd := SearchCmd{}
	cmd.Sort = "relevance"
	cmd.Links = "full"

	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(app.Stderr)
	fs.StringVar(&cmd.Provider, "provider", cmd.Provider, "")
	fs.StringVar(&cmd.Sort, "sort", cmd.Sort, "")
	fs.IntVar(&cmd.YearFrom, "year-from", cmd.YearFrom, "")
	fs.IntVar(&cmd.YearTo, "year-to", cmd.YearTo, "")
	fs.IntVar(&cmd.Limit, "limit", cmd.Limit, "")
	fs.IntVar(&cmd.Offset, "offset", cmd.Offset, "")
	fs.StringVar(&cmd.Format, "format", cmd.Format, "")
	fs.StringVar(&cmd.Links, "links", cmd.Links, "")
	fs.StringVar(&cmd.Seen, "seen", cmd.Seen, "")
	fs.BoolVar(&cmd.NewOnly, "new-only", cmd.NewOnly, "")
	fs.StringVar(&cmd.NewOut, "new-out", cmd.NewOut, "")
	fs.StringVar(&cmd.Out, "out", cmd.Out, "")
	fs.StringVar(&cmd.Out, "output", cmd.Out, "")
	args = reorderFlags(args, map[string]struct{}{
		"provider":  {},
		"sort":      {},
		"year-from": {},
		"year-to":   {},
		"limit":     {},
		"offset":    {},
		"format":    {},
		"links":     {},
		"seen":      {},
		"new-out":   {},
		"out":       {},
		"output":    {},
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := validateSort(cmd.Sort); err != nil {
		return err
	}

	query := strings.Join(fs.Args(), " ")
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("search requires <query>")
	}
	if err := validateSeenFlags(cmd.Seen, cmd.NewOnly, cmd.NewOut); err != nil {
		return err
	}
	cmd.Query = query
	return cmd.Run(app)
}

func dispatchAuthor(app *App, args []string) error {
	if hasHelpFlag(args) {
		printAuthorHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}

	cmd := AuthorCmd{}
	cmd.Sort = "relevance"
	cmd.Links = "full"

	fs := flag.NewFlagSet("author", flag.ContinueOnError)
	fs.SetOutput(app.Stderr)
	fs.StringVar(&cmd.Provider, "provider", cmd.Provider, "")
	fs.StringVar(&cmd.Sort, "sort", cmd.Sort, "")
	fs.IntVar(&cmd.YearFrom, "year-from", cmd.YearFrom, "")
	fs.IntVar(&cmd.YearTo, "year-to", cmd.YearTo, "")
	fs.IntVar(&cmd.Limit, "limit", cmd.Limit, "")
	fs.IntVar(&cmd.Offset, "offset", cmd.Offset, "")
	fs.StringVar(&cmd.Format, "format", cmd.Format, "")
	fs.StringVar(&cmd.Links, "links", cmd.Links, "")
	fs.StringVar(&cmd.Seen, "seen", cmd.Seen, "")
	fs.BoolVar(&cmd.NewOnly, "new-only", cmd.NewOnly, "")
	fs.StringVar(&cmd.NewOut, "new-out", cmd.NewOut, "")
	fs.StringVar(&cmd.Out, "out", cmd.Out, "")
	fs.StringVar(&cmd.Out, "output", cmd.Out, "")
	args = reorderFlags(args, map[string]struct{}{
		"provider":  {},
		"sort":      {},
		"year-from": {},
		"year-to":   {},
		"limit":     {},
		"offset":    {},
		"format":    {},
		"links":     {},
		"seen":      {},
		"new-out":   {},
		"out":       {},
		"output":    {},
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := validateSort(cmd.Sort); err != nil {
		return err
	}

	name := strings.Join(fs.Args(), " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("author requires <name>")
	}
	if err := validateSeenFlags(cmd.Seen, cmd.NewOnly, cmd.NewOut); err != nil {
		return err
	}
	cmd.Name = name
	return cmd.Run(app)
}

func dispatchInfo(app *App, args []string) error {
	if hasHelpFlag(args) {
		printInfoHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}

	cmd := InfoCmd{
		Format: "json",
		Links:  "full",
	}
	fs := flag.NewFlagSet("info", flag.ContinueOnError)
	fs.SetOutput(app.Stderr)
	fs.StringVar(&cmd.Provider, "provider", cmd.Provider, "")
	fs.StringVar(&cmd.Format, "format", cmd.Format, "")
	fs.StringVar(&cmd.Links, "links", cmd.Links, "")
	fs.StringVar(&cmd.Out, "out", cmd.Out, "")
	fs.StringVar(&cmd.Out, "output", cmd.Out, "")
	args = reorderFlags(args, map[string]struct{}{
		"provider": {},
		"format":   {},
		"links":    {},
		"out":      {},
		"output":   {},
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) == 0 {
		return errors.New("info requires <id>")
	}
	cmd.ID = strings.TrimSpace(fs.Args()[0])
	if cmd.ID == "" {
		return errors.New("info requires <id>")
	}
	return cmd.Run(app)
}

func dispatchDownload(app *App, args []string) error {
	if hasHelpFlag(args) {
		printDownloadHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}

	cmd := DownloadCmd{}
	fs := flag.NewFlagSet("download", flag.ContinueOnError)
	fs.SetOutput(app.Stderr)
	fs.StringVar(&cmd.Provider, "provider", cmd.Provider, "")
	fs.StringVar(&cmd.Out, "out", cmd.Out, "")
	fs.StringVar(&cmd.Out, "output", cmd.Out, "")
	fs.StringVar(&cmd.Out, "file", cmd.Out, "")
	args = reorderFlags(args, map[string]struct{}{
		"provider": {},
		"out":      {},
		"output":   {},
		"file":     {},
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) == 0 {
		return errors.New("download requires <id>")
	}
	cmd.ID = strings.TrimSpace(fs.Args()[0])
	if cmd.ID == "" {
		return errors.New("download requires <id>")
	}
	return cmd.Run(app)
}

func dispatchSeen(app *App, args []string) error {
	if hasHelpFlag(args) {
		style := newHelpStyler(app.Stdout, app.Globals)
		if len(args) > 0 {
			switch strings.ToLower(strings.TrimSpace(args[0])) {
			case "diff":
				printSeenDiffHelp(app.Stdout, style)
				return nil
			case "update":
				printSeenUpdateHelp(app.Stdout, style)
				return nil
			}
		}
		printSeenHelp(app.Stdout, style)
		return nil
	}
	if len(args) == 0 {
		return errors.New("seen requires a subcommand: diff | update")
	}
	switch args[0] {
	case "diff":
		return dispatchSeenDiff(app, args[1:])
	case "update":
		return dispatchSeenUpdate(app, args[1:])
	default:
		return fmt.Errorf("unknown seen subcommand %q", args[0])
	}
}

func dispatchSeenDiff(app *App, args []string) error {
	if hasHelpFlag(args) {
		printSeenDiffHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}

	cmd := SeenDiffCmd{}
	fs := flag.NewFlagSet("seen diff", flag.ContinueOnError)
	fs.SetOutput(app.Stderr)
	fs.StringVar(&cmd.New, "new", cmd.New, "")
	fs.StringVar(&cmd.Seen, "seen", cmd.Seen, "")
	fs.StringVar(&cmd.Out, "out", cmd.Out, "")
	fs.StringVar(&cmd.Out, "output", cmd.Out, "")
	fs.StringVar(&cmd.Out, "file", cmd.Out, "")
	fs.BoolVar(&cmd.Stats, "stats", cmd.Stats, "")
	args = reorderFlags(args, map[string]struct{}{
		"new":    {},
		"seen":   {},
		"out":    {},
		"output": {},
		"file":   {},
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(cmd.New) == "" || strings.TrimSpace(cmd.Seen) == "" || strings.TrimSpace(cmd.Out) == "" {
		return errors.New("seen diff requires --new, --seen, and --out")
	}
	return cmd.Run(app)
}

func dispatchSeenUpdate(app *App, args []string) error {
	if hasHelpFlag(args) {
		printSeenUpdateHelp(app.Stdout, newHelpStyler(app.Stdout, app.Globals))
		return nil
	}

	cmd := SeenUpdateCmd{}
	fs := flag.NewFlagSet("seen update", flag.ContinueOnError)
	fs.SetOutput(app.Stderr)
	fs.StringVar(&cmd.Seen, "seen", cmd.Seen, "")
	fs.StringVar(&cmd.Input, "input", cmd.Input, "")
	fs.StringVar(&cmd.Out, "out", cmd.Out, "")
	fs.StringVar(&cmd.Out, "output", cmd.Out, "")
	fs.StringVar(&cmd.Out, "file", cmd.Out, "")
	fs.BoolVar(&cmd.Stats, "stats", cmd.Stats, "")
	args = reorderFlags(args, map[string]struct{}{
		"seen":   {},
		"input":  {},
		"out":    {},
		"output": {},
		"file":   {},
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(cmd.Seen) == "" || strings.TrimSpace(cmd.Input) == "" || strings.TrimSpace(cmd.Out) == "" {
		return errors.New("seen update requires --seen, --input, and --out")
	}
	return cmd.Run(app)
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		if n, err := strconv.Atoi(v); err == nil {
			return n != 0
		}
		return fallback
	}
}

func envString(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func validateSeenFlags(seen string, newOnly bool, newOut string) error {
	if (newOnly || strings.TrimSpace(newOut) != "") && strings.TrimSpace(seen) == "" {
		return errors.New("--new-only and --new-out require --seen")
	}
	return nil
}

func validateColorMode(v string) error {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "auto", "always", "never":
		return nil
	default:
		return fmt.Errorf("invalid value for --color %q (expected auto|always|never)", v)
	}
}

func reorderFlags(args []string, valueFlags map[string]struct{}) []string {
	if len(args) == 0 {
		return nil
	}
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		token := args[i]
		if !strings.HasPrefix(token, "--") || token == "--" {
			positionals = append(positionals, token)
			continue
		}

		name := strings.TrimPrefix(token, "--")
		if idx := strings.Index(name, "="); idx >= 0 {
			flags = append(flags, token)
			continue
		}

		flags = append(flags, token)
		if _, ok := valueFlags[name]; ok && i+1 < len(args) {
			next := args[i+1]
			if !strings.HasPrefix(next, "--") || next == "--" {
				flags = append(flags, next)
				i++
			}
		}
	}

	out := make([]string, 0, len(args))
	out = append(out, flags...)
	out = append(out, positionals...)
	return out
}
