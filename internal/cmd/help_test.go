package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintHelp_GlobalIncludesExamples(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintHelp(&buf, "", nil, Globals{}); err != nil {
		t.Fatalf("PrintHelp returned error: %v", err)
	}

	output := buf.String()
	expected := []string{
		"Examples:",
		"papercli config init",
		"papercli search \"vision transformer\" --provider arxiv --limit 5",
		"papercli info 1706.03762 --provider arxiv --format md",
		"papercli download 1706.03762 --provider arxiv --out attention-is-all-you-need.pdf",
		"papercli seen diff --new latest.json --seen seen.json --out unseen.json --stats",
	}

	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Fatalf("expected help output to contain %q\noutput:\n%s", want, output)
		}
	}
}
