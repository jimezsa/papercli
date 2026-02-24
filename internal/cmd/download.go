package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DownloadCmd struct {
	ID       string `arg:"" required:"" help:"Paper ID."`
	Provider string `help:"Provider to query." enum:"arxiv,semantic,scholar,all" default:"all"`
	Out      string `name:"out" aliases:"output,file" help:"Output PDF path."`
}

func (c *DownloadCmd) Run(app *App) error {
	paper, err := app.Manager.Info(app.Context(), app.EffectiveProvider(c.Provider), c.ID)
	if err != nil {
		return err
	}
	downloadURL := strings.TrimSpace(paper.PDFURL)
	if downloadURL == "" && strings.HasSuffix(strings.ToLower(paper.URL), ".pdf") {
		downloadURL = paper.URL
	}
	if downloadURL == "" {
		return fmt.Errorf("provider %s did not return a downloadable PDF URL", paper.Provider)
	}

	req, err := http.NewRequestWithContext(app.Context(), http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := app.HTTP.Do(app.Context(), req, 45*time.Second, "papercli/0.1")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	path := c.Out
	if strings.TrimSpace(path) == "" {
		path = defaultPDFName(c.ID)
	}
	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", path, err)
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("write PDF %q: %w", path, err)
	}

	_, err = fmt.Fprintf(app.Stdout, "saved %s\n", path)
	return err
}

func defaultPDFName(id string) string {
	clean := strings.TrimSpace(id)
	clean = strings.ReplaceAll(clean, "/", "_")
	clean = strings.ReplaceAll(clean, ":", "_")
	if clean == "" {
		clean = "paper"
	}
	if !strings.HasSuffix(strings.ToLower(clean), ".pdf") {
		clean += ".pdf"
	}
	return filepath.Base(clean)
}
