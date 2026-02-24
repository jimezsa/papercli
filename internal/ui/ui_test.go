package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jimezsa/papercli/internal/models"
)

func TestColorsEnabled(t *testing.T) {
	old, had := os.LookupEnv("NO_COLOR")
	if had {
		if err := os.Unsetenv("NO_COLOR"); err != nil {
			t.Fatalf("unset NO_COLOR: %v", err)
		}
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("NO_COLOR", old)
			return
		}
		_ = os.Unsetenv("NO_COLOR")
	})

	if !ColorsEnabled(ColorAlways, false) {
		t.Fatalf("expected color enabled for --color=always when NO_COLOR is unset")
	}
	if ColorsEnabled(ColorNever, false) != false {
		t.Fatalf("expected color disabled for --color=never")
	}
	if ColorsEnabled(ColorAlways, true) != false {
		t.Fatalf("expected forceDisable to disable color")
	}

	t.Setenv("NO_COLOR", "1")
	if ColorsEnabled(ColorAlways, false) != false {
		t.Fatalf("expected NO_COLOR to disable color")
	}
}

func TestRenderTableShortLinksWithoutHyperlinks(t *testing.T) {
	papers := []models.Paper{
		{
			Provider: models.ProviderArXiv,
			Title:    "A title",
			Authors:  []string{"A. Author"},
			Year:     2025,
			URL:      "https://arxiv.org/abs/1234.5678",
		},
	}

	var buf bytes.Buffer
	if err := RenderTable(&buf, papers, false, LinksShort); err != nil {
		t.Fatalf("RenderTable failed: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected table output, got empty string")
	}
	header := strings.Fields(lines[0])
	wantHeader := []string{"provider", "title", "authors", "year", "url"}
	if len(header) != len(wantHeader) {
		t.Fatalf("unexpected header: %q", lines[0])
	}
	for i := range wantHeader {
		if header[i] != wantHeader[i] {
			t.Fatalf("unexpected header: %q", lines[0])
		}
	}
	if !strings.Contains(out, "https://arxiv.org/abs/1234.5678") {
		t.Fatalf("expected full URL when hyperlinks are unavailable: %q", out)
	}
}

func TestRenderTSVAlwaysWritesFullURL(t *testing.T) {
	papers := []models.Paper{
		{
			ID:       "1234.5678",
			Provider: models.ProviderArXiv,
			Title:    "A title",
			Authors:  []string{"A. Author"},
			Year:     2025,
			URL:      "https://arxiv.org/abs/1234.5678",
		},
	}

	var buf bytes.Buffer
	if err := RenderTSV(&buf, papers, LinksShort); err != nil {
		t.Fatalf("RenderTSV failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "\thttps://arxiv.org/abs/1234.5678\n") {
		t.Fatalf("expected full URL in TSV output: %q", out)
	}
}
