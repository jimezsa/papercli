package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"papercli/internal/models"
)

func WriteJSON(w io.Writer, papers []models.Paper) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(papers)
}

func WriteCSV(w io.Writer, papers []models.Paper) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "provider", "title", "authors", "year", "url"}); err != nil {
		return err
	}
	for _, paper := range papers {
		year := ""
		if paper.Year > 0 {
			year = fmt.Sprintf("%d", paper.Year)
		}
		if err := cw.Write([]string{
			paper.ID,
			string(paper.Provider),
			sanitizeCell(paper.Title),
			strings.Join(paper.Authors, ", "),
			year,
			paper.URL,
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func WriteMarkdown(w io.Writer, papers []models.Paper) error {
	if _, err := fmt.Fprintln(w, "# Papers"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	for _, paper := range papers {
		title := sanitizeCell(paper.Title)
		if paper.URL != "" {
			title = fmt.Sprintf("[%s](%s)", title, paper.URL)
		}
		line := fmt.Sprintf("- %s", title)
		if len(paper.Authors) > 0 {
			line += fmt.Sprintf(" - %s", strings.Join(paper.Authors, ", "))
		}
		if paper.Year > 0 {
			line += fmt.Sprintf(" (%d)", paper.Year)
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func sanitizeCell(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	return v
}
