package ui

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jimezsa/papercli/internal/models"
)

const (
	ColorAuto   = "auto"
	ColorAlways = "always"
	ColorNever  = "never"
)

type LinksMode string

const (
	LinksFull  LinksMode = "full"
	LinksShort LinksMode = "short"
)

func ColorsEnabled(mode string, forceDisable bool) bool {
	if forceDisable {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	switch strings.ToLower(mode) {
	case ColorNever:
		return false
	case ColorAlways:
		return true
	default:
		return isTerminal(os.Stdout) && os.Getenv("TERM") != "dumb"
	}
}

func RenderTable(w io.Writer, papers []models.Paper, color bool, links LinksMode) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "provider\ttitle\tauthors\tyear\turl"); err != nil {
		return err
	}

	for _, paper := range papers {
		provider := string(paper.Provider)
		year := "-"
		if paper.Year > 0 {
			year = fmt.Sprintf("%d", paper.Year)
		}

		if color {
			provider = colorize(provider, "34")
			year = colorize(year, "37")
		}

		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			provider,
			strings.TrimSpace(strings.ReplaceAll(paper.Title, "\n", " ")),
			strings.Join(paper.Authors, ", "),
			year,
			displayURL(paper.URL, links),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func colorize(v, code string) string {
	return "\x1b[" + code + "m" + v + "\x1b[0m"
}

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func RenderTSV(w io.Writer, papers []models.Paper, links LinksMode) error {
	for _, paper := range papers {
		title := strings.TrimSpace(strings.ReplaceAll(paper.Title, "\t", " "))
		title = strings.ReplaceAll(title, "\n", " ")
		authors := strings.Join(paper.Authors, ", ")
		authors = strings.ReplaceAll(authors, "\t", " ")
		year := ""
		if paper.Year > 0 {
			year = fmt.Sprintf("%d", paper.Year)
		}
		if _, err := fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			paper.ID,
			paper.Provider,
			title,
			authors,
			year,
			displayURL(paper.URL, links),
		); err != nil {
			return err
		}
	}
	return nil
}

func displayURL(raw string, mode LinksMode) string {
	if mode != LinksShort {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		return u.Host
	}
	return u.Host + "/" + path
}
