package seen

import (
	"testing"

	"github.com/jimezsa/papercli/internal/models"
)

func TestDiff(t *testing.T) {
	papers := []models.Paper{
		{ID: "a1", Provider: models.ProviderArXiv, URL: "https://arxiv.org/abs/a1"},
		{ID: "b2", Provider: models.ProviderSemantic, URL: "https://semantic.org/b2"},
	}
	store := Store{
		IDs: []string{"arxiv:a1"},
	}

	out := Diff(papers, ToSet(store))
	if len(out) != 1 {
		t.Fatalf("expected 1 unseen paper, got %d", len(out))
	}
	if out[0].ID != "b2" {
		t.Fatalf("expected b2, got %s", out[0].ID)
	}
}

func TestAddPapers(t *testing.T) {
	store := Store{
		IDs: []string{"arxiv:a1"},
	}
	papers := []models.Paper{
		{ID: "a1", Provider: models.ProviderArXiv},
		{ID: "x9", Provider: models.ProviderScholar},
	}

	updated := AddPapers(store, papers)
	if len(updated.IDs) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(updated.IDs))
	}
	if updated.IDs[0] != "arxiv:a1" {
		t.Fatalf("expected first id arxiv:a1, got %q", updated.IDs[0])
	}
	if updated.IDs[1] != "scholar:x9" {
		t.Fatalf("expected second id scholar:x9, got %q", updated.IDs[1])
	}
}
