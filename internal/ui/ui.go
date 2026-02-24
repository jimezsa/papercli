package ui

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jimezsa/papercli/internal/models"
	"github.com/muesli/termenv"
)

const (
	ColorAuto   = "auto"
	ColorAlways = "always"
	ColorNever  = "never"
	LinkColor   = "#87CEEB"
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
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ColorNever:
		return false
	case ColorAlways:
		return true
	default:
		return termenv.NewOutput(os.Stdout).ColorProfile() != termenv.Ascii
	}
}

func RenderTable(w io.Writer, papers []models.Paper, color bool, links LinksMode) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	output := termenv.NewOutput(w)
	hyperlinks := color && isTTY(w)
	if _, err := fmt.Fprintln(tw, "provider\ttitle\tauthors\tyear\turl"); err != nil {
		return err
	}

	for _, paper := range papers {
		year := "-"
		if paper.Year > 0 {
			year = fmt.Sprintf("%d", paper.Year)
		}

		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			string(paper.Provider),
			strings.TrimSpace(strings.ReplaceAll(paper.Title, "\n", " ")),
			strings.Join(paper.Authors, ", "),
			year,
			tableURL(paper.URL, output, color, hyperlinks, links),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func RenderTSV(w io.Writer, papers []models.Paper, links LinksMode) error {
	_ = links
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
			paper.URL,
		); err != nil {
			return err
		}
	}
	return nil
}

func tableURL(raw string, output *termenv.Output, color bool, hyperlinks bool, mode LinksMode) string {
	url := strings.TrimSpace(raw)
	if url == "" {
		return "-"
	}

	display := url
	if mode == LinksShort && hyperlinks {
		display = shortURLLabel(url)
	}

	display = ColorizeLink(output, color, display)
	if hyperlinks {
		display = hyperlink(url, display)
	}
	return display
}

func ColorizeLink(output *termenv.Output, enabled bool, text string) string {
	if !enabled || output == nil {
		return text
	}
	return output.String(text).Foreground(output.Color(LinkColor)).String()
}

func hyperlink(url string, text string) string {
	const esc = "\x1b"
	return esc + "]8;;" + url + esc + "\\" + text + esc + "]8;;" + esc + "\\"
}

func shortURLLabel(raw string) string {
	const maxLen = 60
	label := strings.TrimSpace(raw)
	if parsed, err := url.Parse(raw); err == nil {
		host := strings.TrimPrefix(parsed.Host, "www.")
		if host != "" {
			label = host + parsed.Path
		}
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = raw
	}
	if len(label) > maxLen {
		label = label[:maxLen-3] + "..."
	}
	return label
}
