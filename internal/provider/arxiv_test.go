package provider

import (
	"strings"
	"testing"
)

func TestParseArxivFeed(t *testing.T) {
	xmlInput := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>http://arxiv.org/abs/2501.12345v1</id>
    <updated>2025-01-20T00:00:00Z</updated>
    <published>2025-01-19T00:00:00Z</published>
    <title> Test Paper </title>
    <summary> Some abstract text. </summary>
    <author><name>Jane Doe</name></author>
    <author><name>John Roe</name></author>
    <link rel="alternate" type="text/html" href="https://arxiv.org/abs/2501.12345v1"/>
    <link title="pdf" rel="related" type="application/pdf" href="https://arxiv.org/pdf/2501.12345v1.pdf"/>
  </entry>
</feed>`

	feed, err := parseArxivFeed(strings.NewReader(xmlInput))
	if err != nil {
		t.Fatalf("parseArxivFeed failed: %v", err)
	}
	if len(feed.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(feed.Entries))
	}
	paper, err := arxivEntryToPaper(feed.Entries[0])
	if err != nil {
		t.Fatalf("arxivEntryToPaper failed: %v", err)
	}
	if paper.ID != "2501.12345v1" {
		t.Fatalf("unexpected ID: %q", paper.ID)
	}
	if paper.Year != 2025 {
		t.Fatalf("unexpected year: %d", paper.Year)
	}
	if len(paper.Authors) != 2 {
		t.Fatalf("expected 2 authors, got %d", len(paper.Authors))
	}
	if paper.PDFURL == "" {
		t.Fatalf("expected PDF URL to be set")
	}
}
