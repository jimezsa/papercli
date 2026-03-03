package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jimezsa/papercli/internal/models"
)

type fakeProvider struct {
	name string
	info func(string) (*models.Paper, error)
}

func (f fakeProvider) Name() string { return f.name }

func (f fakeProvider) Search(context.Context, models.SearchParams) ([]models.Paper, error) {
	return nil, nil
}

func (f fakeProvider) Author(context.Context, models.AuthorParams) ([]models.Paper, error) {
	return nil, nil
}

func (f fakeProvider) Info(_ context.Context, id string) (*models.Paper, error) {
	if f.info == nil {
		return nil, errors.New("missing info handler")
	}
	return f.info(id)
}

func TestInfoAllUsesDeterministicProviderOrder(t *testing.T) {
	mgr := NewManager(
		fakeProvider{
			name: "scholar",
			info: func(_ string) (*models.Paper, error) {
				return &models.Paper{ID: "scholar-hit", Provider: models.ProviderScholar}, nil
			},
		},
		fakeProvider{
			name: "semantic",
			info: func(_ string) (*models.Paper, error) {
				return &models.Paper{ID: "semantic-hit", Provider: models.ProviderSemantic}, nil
			},
		},
		fakeProvider{
			name: "arxiv",
			info: func(_ string) (*models.Paper, error) {
				return &models.Paper{ID: "arxiv-hit", Provider: models.ProviderArXiv}, nil
			},
		},
	)

	paper, err := mgr.Info(context.Background(), "all", "paper-id")
	if err != nil {
		t.Fatalf("Info(all) failed: %v", err)
	}
	if paper.ID != "arxiv-hit" {
		t.Fatalf("expected arxiv to win deterministic order, got %q", paper.ID)
	}
}

func TestInfoAllFallsBackByOrder(t *testing.T) {
	mgr := NewManager(
		fakeProvider{
			name: "scholar",
			info: func(_ string) (*models.Paper, error) {
				return &models.Paper{ID: "scholar-hit", Provider: models.ProviderScholar}, nil
			},
		},
		fakeProvider{
			name: "semantic",
			info: func(_ string) (*models.Paper, error) {
				return &models.Paper{ID: "semantic-hit", Provider: models.ProviderSemantic}, nil
			},
		},
		fakeProvider{
			name: "arxiv",
			info: func(_ string) (*models.Paper, error) {
				return nil, errors.New("not found")
			},
		},
	)

	paper, err := mgr.Info(context.Background(), "all", "paper-id")
	if err != nil {
		t.Fatalf("Info(all) failed: %v", err)
	}
	if paper.ID != "semantic-hit" {
		t.Fatalf("expected semantic fallback when arxiv fails, got %q", paper.ID)
	}
}

func TestInfoAllAggregatesErrorsWhenNothingFound(t *testing.T) {
	mgr := NewManager(
		fakeProvider{
			name: "arxiv",
			info: func(_ string) (*models.Paper, error) {
				return nil, errors.New("arxiv error")
			},
		},
		fakeProvider{
			name: "semantic",
			info: func(_ string) (*models.Paper, error) {
				return nil, errors.New("semantic error")
			},
		},
	)

	_, err := mgr.Info(context.Background(), "all", "paper-id")
	if err == nil {
		t.Fatalf("expected Info(all) to fail")
	}
	msg := err.Error()
	if !strings.Contains(msg, "arxiv info") || !strings.Contains(msg, "semantic info") {
		t.Fatalf("expected joined provider errors, got %q", msg)
	}
}
