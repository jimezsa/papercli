package provider

import (
	"context"
	"errors"
	"fmt"

	"papercli/internal/models"
)

var ErrNotConfigured = errors.New("provider not configured")

type Provider interface {
	Name() string
	Search(context.Context, models.SearchParams) ([]models.Paper, error)
	Author(context.Context, models.AuthorParams) ([]models.Paper, error)
	Info(context.Context, string) (*models.Paper, error)
}

func ProviderNames() []string {
	return []string{"arxiv", "semantic", "scholar", "all"}
}

type ProviderError struct {
	Provider string
	Op       string
	Err      error
}

func (e ProviderError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Provider, e.Op, e.Err)
}

func (e ProviderError) Unwrap() error {
	return e.Err
}
