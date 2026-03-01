package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jimezsa/papercli/internal/export"
	"github.com/jimezsa/papercli/internal/models"
	"github.com/jimezsa/papercli/internal/seen"
	"github.com/jimezsa/papercli/internal/ui"
	"github.com/muesli/termenv"
)

type QueryFlags struct {
	Provider string `help:"Provider to query." enum:"arxiv,semantic,scholar,all" default:"all"`
	Sort     string `help:"Sort mode." enum:"relevance,date,citations" default:"relevance"`
	YearFrom int    `name:"year-from" help:"Lower publication year bound."`
	YearTo   int    `name:"year-to" help:"Upper publication year bound."`
	Limit    int    `help:"Maximum number of results."`
	Offset   int    `help:"Result offset." default:"0"`
	Format   string `help:"Output format." enum:"csv,json,md"`
	Links    string `help:"Link rendering mode." enum:"short,full" default:"full"`
	Seen     string `help:"Seen-history JSON file path."`
	NewOnly  bool   `name:"new-only" help:"Only output unseen papers (requires --seen)."`
	NewOut   string `name:"new-out" help:"Always write unseen papers JSON to this file (requires --seen)."`
	Out      string `name:"out" aliases:"output" help:"Output file path. Default is stdout."`
}

func (a *App) RenderPapers(papers []models.Paper, flags QueryFlags) error {
	if flags.NewOnly || flags.NewOut != "" {
		if strings.TrimSpace(flags.Seen) == "" {
			return fmt.Errorf("--new-only and --new-out require --seen")
		}
	}

	if flags.Seen != "" {
		store, err := seen.Load(flags.Seen)
		if err != nil {
			return err
		}
		set := seen.ToSet(store)
		unseen := seen.Diff(papers, set)
		if flags.NewOut != "" {
			if err := seen.SavePapers(flags.NewOut, unseen); err != nil {
				return err
			}
		}
		if flags.NewOnly {
			papers = unseen
		}
	}

	out := a.Stdout
	if flags.Out != "" {
		f, err := os.Create(flags.Out)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", flags.Out, err)
		}
		defer f.Close()
		out = f
	}

	linksMode := ui.LinksFull
	if strings.ToLower(flags.Links) == "short" {
		linksMode = ui.LinksShort
	}

	switch {
	case a.Globals.JSON:
		return export.WriteJSON(out, papers)
	case a.Globals.Plain:
		return ui.RenderTSV(out, papers, linksMode)
	}

	switch strings.ToLower(strings.TrimSpace(flags.Format)) {
	case "json":
		return export.WriteJSON(out, papers)
	case "md":
		return export.WriteMarkdown(out, papers)
	case "csv":
		return export.WriteCSV(out, papers)
	case "":
		if flags.Out != "" {
			return export.WriteCSV(out, papers)
		}
		if !isTTY(out) {
			return export.WriteCSV(out, papers)
		}
		color := ui.ColorsEnabled(a.Globals.Color, a.Globals.JSON || a.Globals.Plain)
		return ui.RenderTable(out, papers, color, linksMode)
	default:
		return fmt.Errorf("unsupported format %q", flags.Format)
	}
}

func isTTY(out io.Writer) bool {
	return termenv.NewOutput(out).ColorProfile() != termenv.Ascii
}
