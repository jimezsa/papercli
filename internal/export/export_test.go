package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jimezsa/papercli/internal/models"
)

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	papers := []models.Paper{
		{
			ID:       "1234.5678",
			Provider: models.ProviderArXiv,
			Title:    "A Paper",
			Authors:  []string{"A. Author"},
			Year:     2025,
			URL:      "https://arxiv.org/abs/1234.5678",
		},
	}
	if err := WriteCSV(&buf, papers); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "id,provider,title,authors,year,url") {
		t.Fatalf("missing header row: %q", out)
	}
	if !strings.Contains(out, "1234.5678,arxiv,A Paper,A. Author,2025,https://arxiv.org/abs/1234.5678") {
		t.Fatalf("missing paper row: %q", out)
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	papers := []models.Paper{
		{
			ID:       "p1",
			Provider: models.ProviderSemantic,
			Title:    "Semantic Paper",
			Authors:  []string{"J. Doe", "R. Roe"},
			Year:     2024,
			URL:      "https://example.com/p1",
		},
	}
	if err := WriteMarkdown(&buf, papers); err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Papers") {
		t.Fatalf("missing title: %q", out)
	}
	if !strings.Contains(out, "[Semantic Paper](https://example.com/p1)") {
		t.Fatalf("missing linked title: %q", out)
	}
}
